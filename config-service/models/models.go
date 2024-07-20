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
		DatabaseEntry: DatabaseEntry{},
		ServerEntry:   ServerEntry{},
	}
}

type Models struct {
	DatabaseEntry DatabaseEntry
	ServerEntry   ServerEntry
}

type DatabaseEntry struct {
	Name          string        `bson:"name" json:"name"`
	Password      string        `bson:"password" json:"password"`
	Server        string        `bson:"server" json:"server"`
	Environment   string        `bson:"environment" json:"environment"`
	Configuration Configuration `bson:"configuration" json:"configuration"`
	Connectivity  string        `bson:"connectivity" json:"connectivity"`
	Type          string        `bson:"type" json:"type"`
	Version       string        `bson:"version" json:"version"`
	NodeIP        string        `bson:"node_ip" json:"node_ip"`
	NodePort      string        `bson:"node_port" json:"node_port"`
	DirectoryUUID string        `bson:"directory_uuid" json:"directory_uuid"`
	GrafanaUID    string        `bson:"grafana_uid" json:"grafana_uid"`
	Email         string        `bson:"email" json:"email"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	Status        string        `bson:"status" json:"status"`
}

type Configuration struct {
	ServiceType     string `bson:"service_type" json:"service_type"`
	ComputeType     string `bson:"compute_type" json:"compute_type"`
	MaxStorageSize  string `bson:"max_storage_size" json:"max_storage_size"`
	StorageSizeUnit string `bson:"storage_size_unit" json:"storage_size_unit"`
}

type ServerEntry struct {
	Name      string    `bson:"name" json:"name"`
	Location  string    `bson:"location" json:"location"`
	Admin     string    `bson:"admin" json:"admin"`
	Password  string    `bson:"password" json:"password"`
	Email     string    `bson:"email" json:"email"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Status    string    `bson:"status" json:"status"`
}

func (s *ServerEntry) Insert(entry ServerEntry) error {

	collection := client.Database(DBName).Collection("server")

	_, err := collection.InsertOne(context.TODO(), ServerEntry{
		Name:      entry.Name,
		Location:  entry.Location,
		Admin:     entry.Admin,
		Password:  entry.Password,
		Email:     entry.Email,
		CreatedAt: entry.CreatedAt,
		Status:    entry.Status,
	})

	if err != nil {
		log.Println("Error inserting server entry. Error: ", err)
		return err
	}

	return nil
}

func (d *DatabaseEntry) UpdateGrafanaUID(directoryUUID string, grafanaUID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("database")

	filter := bson.M{"directory_uuid": directoryUUID}
	update := bson.M{
		"$set": bson.M{
			"grafana_uid": grafanaUID,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error updating Grafana UID. Error: ", err)
		return err
	}

	return nil
}

func (d *DatabaseEntry) Insert(entry DatabaseEntry) error {

	collection := client.Database(DBName).Collection("database")

	_, err := collection.InsertOne(context.TODO(), DatabaseEntry{
		Name:          entry.Name,
		Password:      entry.Password,
		Server:        entry.Server,
		Environment:   entry.Environment,
		Configuration: entry.Configuration,
		Connectivity:  entry.Connectivity,
		Type:          entry.Type,
		Version:       entry.Version,
		NodeIP:        entry.NodeIP,
		NodePort:      entry.NodePort,
		DirectoryUUID: entry.DirectoryUUID,
		GrafanaUID:    entry.GrafanaUID,
		Email:         entry.Email,
		CreatedAt:     entry.CreatedAt,
		Status:        entry.Status,
	})

	if err != nil {
		log.Println("Error inserting database entry. Error: ", err)
		return err
	}

	return nil
}

func (s *ServerEntry) GetOne(name string) (*ServerEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("server")

	filter := bson.M{"name": name}

	var entry ServerEntry
	err := collection.FindOne(ctx, filter).Decode(&entry)

	if err != nil {
		log.Println("Error getting server entry. Error: ", err)
		return nil, err
	}

	return &entry, nil
}

func (d *DatabaseEntry) GetOne(name string, email string) (*DatabaseEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("database")

	filter := bson.M{"name": name, "email": email}

	var entry DatabaseEntry
	err := collection.FindOne(ctx, filter).Decode(&entry)
	if err != nil {
		log.Println("Error getting database entry. Error: ", err)
		return nil, err
	}

	return &entry, nil
}

func (d *DatabaseEntry) GetAllByEmail(email string) ([]*DatabaseEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("database")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.M{"email": email}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println("Error getting database entries. Error: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*DatabaseEntry

	for cursor.Next(ctx) {
		var entry DatabaseEntry

		err = cursor.Decode(&entry)
		if err != nil {
			log.Println("Error decoding database entry. Error: ", err)
			return nil, err
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (s *ServerEntry) GetAll(email string) ([]*ServerEntry, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database(DBName).Collection("server")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.M{"email": email}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println("Error getting server entries. Error: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*ServerEntry

	for cursor.Next(ctx) {
		var entry ServerEntry

		err = cursor.Decode(&entry)
		if err != nil {
			log.Println("Error decoding server entry. Error: ", err)
			return nil, err
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}
