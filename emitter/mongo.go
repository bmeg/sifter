
package emitter

import (
  "time"
  "context"
  "github.com/bmeg/grip/gripql"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/bmeg/grip/protoutil"
)


type MongoEmitter struct {
  edgeCol *mongo.Collection
  vertexCol *mongo.Collection
  edgeChan  chan bson.M
  vertexChan  chan bson.M
}

var batchSize int = 100

// NewMongoEmitter
// url : "mongodb://localhost:27017"
func NewMongoEmitter(uri string) (MongoEmitter, error) {
  client, err := mongo.NewClient(options.Client().ApplyURI(uri))
  if err != nil {
    return MongoEmitter{}, err
  }
  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
  err = client.Connect(ctx)

  edgeCol := client.Database("grip").Collection("database_edges")
  vertexCol := client.Database("grip").Collection("database_vertices")

  edgeChan := make(chan bson.M, 100)
  vertexChan := make(chan bson.M, 100)

  return MongoEmitter{edgeCol, vertexCol, edgeChan, vertexChan}, nil
}

func docWriter(col *mongo.Collection, docChan chan bson.M) {
  docBatch := make([]mongo.WriteModel, batchSize)
  for ent := range docChan {
    i := mongo.NewInsertOneModel().SetDocument(ent)
    docBatch = append(docBatch, i)
    if len(docBatch) > batchSize {
      col.BulkWrite(context.Background(), docBatch)
      docBatch = make([]mongo.WriteModel, batchSize)
    }
  }
  if len(docBatch) > 0 {
    col.BulkWrite(context.Background(), docBatch)
  }
}

func (s MongoEmitter) EmitVertex(v *gripql.Vertex) error {
  s.vertexChan <- packVertex(v)
  return nil
}

func (s MongoEmitter) EmitEdge(e *gripql.Edge) error {
  s.edgeChan <- packEdge(e)
  return nil
}


// these are copied in from grip, because that codebase is still linked to older
// mongo driver

func packVertex(v *gripql.Vertex) bson.M {
	p := map[string]interface{}{}
	if v.Data != nil {
		p = protoutil.AsMap(v.Data)
	}
	return bson.M {
		"_id":   v.Gid,
		"label": v.Label,
		"data":  p,
	}
}

func packEdge(e *gripql.Edge) bson.M {
	p := map[string]interface{}{}
	if e.Data != nil {
		p = protoutil.AsMap(e.Data)
	}
	return bson.M {
		"_id":   e.Gid,
		"from":  e.From,
		"to":    e.To,
		"label": e.Label,
		"data":  p,
	}
}
