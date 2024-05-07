package adi

import (
	"bytes"
	"strings"
	"context"
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
	Hosts	   []string	      `bson:"hosts"`
	Vars       map[string]interface{}
}

type HostVars struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Vars       map[string]interface{}
}

type Data struct {
	Groups	[]Group
	Hosts	[]HostVars
}

// MongoDB connection struct
type MongoDB struct {
    Client *mongo.Client
    Ctx    context.Context
    DB     *mongo.Database

    // Add fields for collections here
    Collection *mongo.Collection
    // Add more collections if needed
}

// MDb creates a new MongoDB connection
func MDb(ctx context.Context, uri, dbName string, collection string) (*MongoDB, error) {
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
        Collection: myCollection,
        // Assign more collections here
    }, nil
}

// Close closes the MongoDB connection
func (db *MongoDB) Close() error {
    return db.Client.Disconnect(db.Ctx)
}

func (db *MongoDB) Getgroups(group string, host bool) ([]Group, error) {
    var results []Group
    filter := bson.M{}
    if host {
        filter = bson.M{"hosts": primitive.Regex{Pattern: group, Options: "i"}}
    } else {
        filter = bson.M{"name": group}
    }
    cursor, err := db.Collection.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())
    for cursor.Next(context.Background()) {
        var result Group
        err := cursor.Decode(&result)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    if err := cursor.Err(); err != nil {
        return nil, err
    }
    return results, nil
}

func (db *MongoDB) Gethosts(host string, collection string) ([]HostVars, error) {
    var results []HostVars
    hosts := db.DB.Collection(collection)
    // db.Collection = db.Collection(collection)
    filter := bson.M{"name": primitive.Regex{Pattern: host, Options: "i"}}
    cursor, err := hosts.Find(context.Background(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())
    for cursor.Next(context.Background()) {
        var result HostVars
        err := cursor.Decode(&result)
        if err != nil {
            return nil, err
        }
        results = append(results, result)
    }
    if err := cursor.Err(); err != nil {
        return nil, err
    }
    return results, nil
}

func (db *MongoDB) Hostgroup (group string, dc string) ([]string, error) {
    var results []Group
    filter := bson.M{"name": group, "datacenter": dc}
    cursor, err := db.Collection.Find(context.Background(), filter)
    if err != nil {
        return []string{}, err
    }
    defer cursor.Close(context.Background())
    for cursor.Next(context.Background()) {
        var result Group
        err := cursor.Decode(&result)
        if err != nil {
            return []string{}, err
        }
        results = append(results, result)
    }
    if err := cursor.Err(); err != nil {
        return []string{}, err
    }

    if (len(results) == 1) {
        return results[0].Hosts, nil
    }
    return []string{}, nil
}

func (data Data) GenerateJSON() (bytes.Buffer, error) {
	// Define the JSON template
	jsonTemplate := `{
	{{range $index, $Groups := .Groups }}"{{ $Groups.Name }}": {
		{{if len $Groups.Children }}"children" : [
			{{range $i, $member := $Groups.Children}}"{{ $member }}"{{if notLastCommaChild $i $Groups.Children}},{{end}}{{end}}
		],{{end}}
		{{if len $Groups.Hosts }}"hosts": [
			{{range $i, $member := $Groups.Hosts}}"{{ $member }}"{{if notLastCommaChild $i $Groups.Hosts}},{{end}}{{end}}
		],{{end}}
		{{with $Groups.Vars }}"vars" : {{  jsonify $Groups.Vars }},{{end}}
		"datacenter" : "{{ $Groups.Datacenter }}"
	},{{end}}
	"_meta": {
		"hostvars": {
			{{range $index, $Hosts := .Hosts }}"{{ $Hosts.Name }}":{{  jsonify $Hosts.Vars }}{{if notLastCommaHost $index}},{{end}}
			{{end}}
		}
	}
}`

	// Define custom function for checking if it's the last iteration
	funcMap := template.FuncMap{
		"notLastComma": func(index int) bool {
			return index != len(data.Groups)-1
		},
		"notLastCommaHost": func(index int) bool {
			return index != len(data.Hosts)-1
		},
		"notLastCommaChild": func(index int, members []string) bool {
			return index != len(members)-1
		},
		"jsonify": func(v interface{}) (string, error) {
			jsonData, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
	}

    var jsonData bytes.Buffer
	// Create a new template and parse the template string
	tmpl, err := template.New("jsonTemplate").Funcs(funcMap).Parse(jsonTemplate)
	if err != nil {
        return jsonData, err
	}

	err = tmpl.Execute(&jsonData, data)
	if err != nil {
        return jsonData, err
	}

	// Return the result as a string
	return jsonData, nil
}

func Validate(input []string) string {
	// Create a map to store unique values
	unique := make(map[string]bool)
	for _, str := range input {
		// Find the last occurrence of "."
        parts := strings.Split(str, ".")
        var dom string
        if len(parts) >= 3 {
            dom = strings.Join(parts[len(parts)-3:], ".")
        }
    
        br := strings.Split(str, dom)
        clus := strings.Split(br[0], "br")
        unique[clus[0]] = true

	}

	// Convert the unique keys from the map to a slice
	var validated []string
	for key := range unique {
		validated = append(validated, key)
	}
	return validated[0]
}
