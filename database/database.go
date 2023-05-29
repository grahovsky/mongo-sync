package database

import (
	"context"

	"github.com/grahovsky/mongo-sync/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Doc struct {
	document   bson.M
	collection string
}

type Conn struct {
	ConnString string
	Username   string
	Password   string
	AuthSourc  string
}

func GetClient(c Conn) *mongo.Client {
	logger := common.GetLogger()

	clientOptions := options.Client().ApplyURI(c.ConnString).SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      c.Username,
		Password:      c.Password,
		AuthSource:    c.AuthSourc,
	})

	clientDst, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logger.Fatal(err)
	}

	return clientDst
}
