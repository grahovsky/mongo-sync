package db

import (
	"context"
	"log/slog"
	"sync"

	"mongo-sync/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FindReq struct {
	Collection string
	Filter     bson.M
}

type SaveReq struct {
	Collection string
	Data       []interface{}
}

type (
	MongoClients map[string]*mongo.Client

	MongoPull struct {
		mu      sync.RWMutex
		Clients MongoClients
	}
)

var Pull MongoPull

func init() {
	Pull.Clients = make(MongoClients)
}

func NewClient(conn config.DBConn) (*mongo.Client, error) {
	Pull.mu.Lock()
	defer Pull.mu.Unlock()

	ctx := context.Background()

	if cached, ok := Pull.Clients[conn.URI+conn.Name]; ok {
		slog.Debug("trying to find cached client..")
		if err := cached.Ping(ctx, nil); err == nil {
			slog.Debug("cached client found")
			return cached, nil
		}
		slog.Debug("cached client not found")
	}

	clientOptions := options.Client().ApplyURI(conn.URI)
	clientOptions.SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      conn.Auth.Username,
		Password:      conn.Auth.Password,
		AuthSource:    conn.Auth.Source,
	})
	clientOptions.SetTimeout(conn.Timeout)

	newClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	if err := newClient.Ping(ctx, nil); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	Pull.Clients[conn.URI+conn.Name] = newClient
	return newClient, nil
}
