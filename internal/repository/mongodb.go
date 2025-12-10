package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/levskiy0/m3m/internal/config"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Select URI and database based on driver configuration
	var uri, dbName string
	switch cfg.Database.Driver {
	case "ferretdb":
		uri = cfg.FerretDB.URI
		dbName = cfg.FerretDB.Database
	default: // "mongodb" or empty
		uri = cfg.MongoDB.URI
		dbName = cfg.MongoDB.Database
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(dbName)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
