
package emitter

import (
  "time"
  "context"
  "sync"
  "log"
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
  writers    *sync.WaitGroup
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

  s := &sync.WaitGroup{}
  go docWriter(edgeCol, edgeChan, s)
  s.Add(1)
  go docWriter(vertexCol, vertexChan, s)
  s.Add(1)

  return MongoEmitter{edgeCol, vertexCol, edgeChan, vertexChan, s}, nil
}

/*
func addGraph(client mongo.Client) {
	graphs := session.DB(ma.database).C("graphs")
	err = graphs.Insert(bson.M{"_id": graph})
	if err != nil {
		return fmt.Errorf("failed to insert graph %s: %v", graph, err)
	}


  	e := ma.EdgeCollection(session, graph)
  	err := e.EnsureIndex(mgo.Index{
  		Key:        []string{"from"},
  		Unique:     false,
  		DropDups:   false,
  		Sparse:     false,
  		Background: true,
  	})
  	if err != nil {
  		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
  	}
  	err = e.EnsureIndex(mgo.Index{
  		Key:        []string{"to"},
  		Unique:     false,
  		DropDups:   false,
  		Sparse:     false,
  		Background: true,
  	})
  	if err != nil {
  		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
  	}
  	err = e.EnsureIndex(mgo.Index{
  		Key:        []string{"label"},
  		Unique:     false,
  		DropDups:   false,
  		Sparse:     false,
  		Background: true,
  	})
  	if err != nil {
  		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
  	}

  	v := ma.VertexCollection(session, graph)
  	err = v.EnsureIndex(mgo.Index{
  		Key:        []string{"label"},
  		Unique:     false,
  		DropDups:   false,
  		Sparse:     false,
  		Background: true,
  	})
  	if err != nil {
  		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
  	}

  	return nil

}
*/

func docWriter(col *mongo.Collection, docChan chan bson.M, sn *sync.WaitGroup) {
  defer sn.Done()
  docBatch := make([]mongo.WriteModel, 0, batchSize)
  for ent := range docChan {
    i := mongo.NewInsertOneModel()
    i.SetDocument(ent)
    docBatch = append(docBatch, i)
    if len(docBatch) > batchSize {
      _, err := col.BulkWrite(context.Background(), docBatch)
      if err != nil {
        log.Printf("%s", err)
      }
      docBatch = make([]mongo.WriteModel, 0, batchSize)
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

func (s MongoEmitter) Close() {
  close(s.vertexChan)
  close(s.edgeChan)
  s.writers.Wait()
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
