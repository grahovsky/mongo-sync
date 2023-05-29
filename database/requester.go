package database

import (
	"context"
	"time"

	"github.com/grahovsky/mongo-sync/common"
	"go.mongodb.org/mongo-driver/bson"
)

func PrepareSend(config common.Config, jobs chan<- Doc, requestTime time.Time) {
	logger := common.GetLogger()

	// Connect to source database
	conn := Conn{
		config.SrcConnStr,
		config.SrcUsername,
		config.SrcPassword,
		config.SrcAuthsorce,
	}
	clientSrc := GetClient(conn)
	defer func() {
		errDisc := clientSrc.Disconnect(context.Background())
		if errDisc != nil {
			logger.Fatal(errDisc)
		}
	}()

	dbSrc := clientSrc.Database(config.SrcDbName)

	logger.Println("Start sync iteration")

	// Getting a list of collections in the source base
	collectionNames, err := dbSrc.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		logger.Fatal(err)
	}

	// fetch new documents from all collections of the source base
	for _, collectionName := range collectionNames {
		collectionSrc := dbSrc.Collection(collectionName)

		filter := bson.M{config.SyncDateField: bson.M{"$gt": requestTime}}
		cur, err := collectionSrc.Find(context.Background(), filter)
		if err != nil {
			logger.Fatal(err)
		}

		for cur.Next(context.Background()) {

			var document bson.M
			if err := cur.Decode(&document); err != nil {
				logger.Fatal(err)
			}
			jobs <- Doc{document: document, collection: collectionName}

			logger.Println("finded", document)
		}

		if err := cur.Err(); err != nil {
			logger.Fatal(err)
		}
		cur.Close(context.Background())
	}
}
