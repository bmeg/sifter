package datastore

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config describes the configuration for the mongodb driver.
type Config struct {
	URL        string
	Database   string
	Collection string
	Username   string
	Password   string
	BatchSize  int
}

type DataStore interface {
	HasRecordStream(key string) bool
	GetRecordStream(key string) chan map[string]interface{}
	SetRecordStream(key string, data chan map[string]interface{}) error
}

type mongoStore struct {
	client     *mongo.Client
	database   string
	collection string
}

func GetMongoStore(conf Config) (DataStore, error) {
	log.Printf("Starting Mongo DataStore connection")

	clientOpts := options.Client()
	clientOpts.SetAppName("sifter")
	clientOpts.SetConnectTimeout(1 * time.Minute)
	if conf.Username != "" || conf.Password != "" {
		cred := options.Credential{Username: conf.Username, Password: conf.Password}
		clientOpts.SetAuth(cred)
	}
	clientOpts.SetRetryReads(true)
	clientOpts.SetRetryWrites(true)
	clientOpts.SetMaxPoolSize(4096)
	clientOpts.SetMaxConnIdleTime(10 * time.Minute)
	clientOpts.ApplyURI(conf.URL)

	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return mongoStore{client: client, database: conf.Database, collection: conf.Collection}, nil
}

func (ms mongoStore) SetRecordStream(key string, data chan map[string]interface{}) error {
	coll := ms.client.Database(ms.database).Collection(ms.collection)
	_, err := coll.InsertOne(context.Background(), bson.M{"_id": key})
	if err != nil {
		log.Printf("Cache Insert Error: %s", err)
	}
	conColl := ms.client.Database(ms.database).Collection(ms.collection + "_contents")
	for i := range data {
		tmp := bson.M{}
		for k, v := range i {
			tmp[k] = v
		}
		tmp["_srcid"] = key
		_, err := conColl.InsertOne(context.Background(), tmp)
		if err != nil {
			log.Printf("Cache Insert Error: %s", err)
		}
	}
	return nil
}

func (ms mongoStore) HasRecordStream(key string) bool {
	coll := ms.client.Database(ms.database).Collection(ms.collection)
	result := coll.FindOne(context.Background(), bson.M{"_id": key})
	if result.Err() != nil {
		return false
	}
	return true
}

func (ms mongoStore) GetRecordStream(key string) chan map[string]interface{} {
	coll := ms.client.Database(ms.database).Collection(ms.collection + "_contents")

	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(out)
		cursor, err := coll.Find(context.TODO(), bson.M{"_srcid": key})
		if err != nil {
			log.Printf("Errors: %s", err)
		}
		result := map[string]interface{}{}
		for cursor.Next(context.TODO()) {
			if nil == cursor.Decode(&result) {
				out <- result
			}
		}
		if err := cursor.Close(context.TODO()); err != nil {
			log.Printf("MongoDB: Record stream error")
		}
	}()
	return out
}
