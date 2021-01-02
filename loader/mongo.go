package loader

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/protoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bmeg/sifter/schema"
)

type MongoLoader struct {
	edgeCol    *mongo.Collection
	vertexCol  *mongo.Collection
	edgeChan   chan bson.M
	vertexChan chan bson.M
	writers    *sync.WaitGroup
}

var batchSize int = 100
var database string = "gripdb"

func MongoGraphExists(uri string, graph string) (bool, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return false, err
	}

	graphs := client.Database(database).Collection("graphs")
	cur, err := graphs.Find(ctx, bson.M{"_id": graph})
	if err != nil {
		log.Printf("Graph find failed: %s", err)
		return false, err
	}
	out := false
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		out = true
	}
	return out, nil
}

// NewMongoEmitter
// url : "mongodb://localhost:27017"
func NewMongoLoader(uri string, graph string) (MongoLoader, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return MongoLoader{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)

	addGraph(client, database, graph)

	edgeCol := edgeCollection(client, database, graph)
	vertexCol := vertexCollection(client, database, graph)

	edgeChan := make(chan bson.M, 100)
	vertexChan := make(chan bson.M, 100)

	s := &sync.WaitGroup{}
	go docWriter(edgeCol, edgeChan, s)
	s.Add(1)
	go docWriter(vertexCol, vertexChan, s)
	s.Add(1)

	return MongoLoader{edgeCol, vertexCol, edgeChan, vertexChan, s}, nil
}

func (s MongoLoader) NewDataEmitter(sc *schema.Schemas) (DataEmitter, error) {
	return nil, fmt.Errorf("Mongo data loader not implemented")
}

func (s MongoLoader) NewGraphEmitter() (GraphEmitter, error) {
	return s, nil
}

func boolPtr(a bool) *bool {
	return &a
}

func addGraph(client *mongo.Client, database string, graph string) error {
	graphs := client.Database(database).Collection("graphs")
	_, err := graphs.InsertOne(context.Background(), bson.M{"_id": graph})
	if err != nil {
		return fmt.Errorf("failed to insert graph %s: %v", graph, err)
	}

	e := edgeCollection(client, database, graph)
	eiv := e.Indexes()
	_, err = eiv.CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: []string{"from"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	_, err = eiv.CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: []string{"to"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	_, err = eiv.CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: []string{"label"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	v := vertexCollection(client, database, graph)
	viv := v.Indexes()
	_, err = viv.CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: []string{"label"},
			Options: &options.IndexOptions{
				Unique:     boolPtr(false),
				Sparse:     boolPtr(false),
				Background: boolPtr(true),
			},
		})
	if err != nil {
		return fmt.Errorf("failed create index for graph %s: %v", graph, err)
	}

	return nil

}

func vertexCollection(session *mongo.Client, database string, graph string) *mongo.Collection {
	return session.Database(database).Collection(fmt.Sprintf("%s_vertices", graph))
}

func edgeCollection(session *mongo.Client, database string, graph string) *mongo.Collection {
	return session.Database(database).Collection(fmt.Sprintf("%s_edges", graph))
}

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

func (s MongoLoader) EmitVertex(v *gripql.Vertex) error {
	s.vertexChan <- packVertex(v)
	return nil
}

func (s MongoLoader) EmitEdge(e *gripql.Edge) error {
	s.edgeChan <- packEdge(e)
	return nil
}

func (s MongoLoader) Close() {
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
	return bson.M{
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
	return bson.M{
		"_id":   e.Gid,
		"from":  e.From,
		"to":    e.To,
		"label": e.Label,
		"data":  p,
	}
}
