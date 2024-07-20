package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var DBName string
var DB Models

func New(DBNameP string, clientP *mongo.Client) {
	DBName = DBNameP
	client = clientP

	DB = Models{
		RegionEntry: RegionEntry{},
	}
}

type Models struct {
	RegionEntry RegionEntry
}

type RegionEntry struct {
	Name  string `bson:"name" json:"name"`
	Types []Type
}

type Type struct {
	Name     string `bson:"name" json:"name"`
	Versions []Version
}

type Version struct {
	Name string `bson:"name" json:"name"`
}

func (r *RegionEntry) Insert(entry RegionEntry) error {

	collection := client.Database(DBName).Collection("region")

	_, err := collection.InsertOne(context.TODO(), RegionEntry{
		Name:  entry.Name,
		Types: []Type{},
	})

	if err != nil {
		log.Println("Error inserting region entry. Error: ", err)
		return err
	}

	return nil
}

func (r *RegionEntry) GetOne(name string) (*RegionEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("region")

	filter := bson.M{"name": name}

	var entry RegionEntry
	err := collection.FindOne(ctx, filter).Decode(&entry)

	if err != nil {
		log.Println("Error getting region entry. Error: ", err)
		return nil, err
	}

	return &entry, nil
}

func (r *RegionEntry) GetAll() ([]RegionEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("region")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "name", Value: 1}})
	filter := bson.D{}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println("Error getting region entries. Error: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []RegionEntry

	for cursor.Next(ctx) {
		var entry RegionEntry

		err = cursor.Decode(&entry)
		if err != nil {
			log.Println("Error decoding region entry. Error: ", err)
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *RegionEntry) AddType(region string, newType Type) error {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("region")

	filter := bson.M{"name": region}

	update := bson.M{
		"$push": bson.M{
			"types": bson.M{
				"name":     newType.Name,
				"versions": []Version{},
			},
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update, nil)

	if err != nil {
		log.Println("Error adding type. Error: ", err)
		return err
	}

	return nil
}

func (r *RegionEntry) AddVersion(region string, dbType string, version Version) error {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("region")

	filter := bson.M{
		"name":       region,
		"types.name": dbType,
	}

	update := bson.M{
		"$push": bson.M{
			"types.$[elem].versions": bson.M{
				"name": version.Name,
			},
		},
	}

	arrFilter := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.name": dbType},
		},
	})

	_, err := collection.UpdateOne(ctx, filter, update, arrFilter)

	if err != nil {
		log.Println("Error adding version. Error: ", err)
		return err
	}

	return nil
}
