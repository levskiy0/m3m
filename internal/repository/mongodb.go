package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/FerretDB/FerretDB/ferretdb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/levskiy0/m3m/internal/config"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	ferret   *ferretdb.FerretDB // embedded FerretDB instance (nil for pure MongoDB)
	cancel   context.CancelFunc // cancel function for embedded FerretDB
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var uri, dbName string
	var ferret *ferretdb.FerretDB
	var ferretCancel context.CancelFunc

	switch cfg.Database.Driver {
	case "sqlite":
		// Ensure data directory exists
		dataPath := cfg.SQLite.Path
		if err := os.MkdirAll(dataPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create SQLite data directory: %w", err)
		}

		absPath, err := filepath.Abs(dataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Create embedded FerretDB with SQLite backend
		f, err := ferretdb.New(&ferretdb.Config{
			Listener: ferretdb.ListenerConfig{
				TCP: "127.0.0.1:0", // Use random available port
			},
			Handler:   "sqlite",
			SQLiteURL: "file:" + absPath + "/",
			Logger:    slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create embedded FerretDB: %w", err)
		}

		// Run FerretDB in background
		ferretCtx, fCancel := context.WithCancel(context.Background())
		ferretCancel = fCancel
		go f.Run(ferretCtx)

		// Wait a bit for FerretDB to start
		time.Sleep(100 * time.Millisecond)

		ferret = f
		uri = f.MongoDBURI()
		dbName = cfg.SQLite.Database

	default: // "mongodb" or empty
		uri = cfg.MongoDB.URI
		dbName = cfg.MongoDB.Database
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		if ferretCancel != nil {
			ferretCancel()
		}
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		if ferretCancel != nil {
			ferretCancel()
		}
		return nil, err
	}

	database := client.Database(dbName)

	return &MongoDB{
		Client:   client,
		Database: database,
		ferret:   ferret,
		cancel:   ferretCancel,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.Client.Disconnect(ctx)

	// Stop embedded FerretDB if running
	if m.cancel != nil {
		m.cancel()
	}

	return err
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
