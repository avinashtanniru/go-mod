package adi

import (
	"strings"
	"fmt"
	"log"
	"bytes"
	"context"
	"net/http"
	"encoding/json"
	"text/template"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Children   []string           `bson:"children"`
	Datacenter string             `bson:"datacenter"`
	Name       string             `bson:"name"`
	Hosts	   []string			  `bson:"hosts"`
	Vars       map[string]interface{}
}

type HostVars struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Vars       map[string]interface{}
}

type Data struct {
	Groups	[]AGroup
	Hosts	[]HostVars
}

// MongoDB connection struct
type MongoDB struct {
    Client *mongo.Client
    Ctx    context.Context
    DB     *mongo.Database

    // Add fields for collections here
    MyCollection *mongo.Collection
    // Add more collections if needed
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(ctx context.Context, uri, dbName string, collection string) (*MongoDB, error) {
    client, err := mongo.NewClient(options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }

    // Connect to MongoDB
    err = client.Connect(ctx)
    if err != nil {
        return nil, err
    }

    // Check the connection
    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, err
    }

    db := client.Database(dbName)

    // Initialize collections
    myCollection := db.Collection(collection)
    // Initialize more collections if needed

    return &MongoDB{
        Client:       client,
        Ctx:          ctx,
        DB:           db,
        MyCollection: myCollection,
        // Assign more collections here
    }, nil
}

// Close closes the MongoDB connection
func (db *MongoDB) Close() error {
    return db.Client.Disconnect(db.Ctx)
}


// var AGroups = []string{
//     "mongodb-hosts",
//     "mongodb-vzs",
// }
