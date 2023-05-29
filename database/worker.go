package database

import (
	"context"
	"sync"

	"github.com/grahovsky/mongo-sync/common"
	"go.mongodb.org/mongo-driver/bson"
)

func Worker(jobs <-chan Doc, config *common.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := common.GetLogger()

	// Connecting to a Destination Base
	conn := Conn{
		config.DstConnStr,
		config.DstUsername,
		config.DstPassword,
		config.DstAuthsorce,
	}
	clientDst := GetClient(conn)
	defer func() {
		clientDst := clientDst
		err := clientDst.Disconnect(context.Background())
		if err != nil {
			logger.Fatal(err)
		}
	}()

	dbDst := clientDst.Database(config.DstDbName)

	for doc := range jobs {

		document := doc.document
		collectionDst := dbDst.Collection(doc.collection)

		// look for documents in the destination database by _id
		count, err := collectionDst.CountDocuments(context.Background(), bson.M{"_id": document["_id"]})
		if err != nil {
			logger.Fatal(err)
		}

		// if there is no document, add them
		if count == 0 {
			logger.Println("insert", document)
			_, err := collectionDst.InsertOne(context.Background(), document)
			if err != nil {
				logger.Fatal(err)
			}
			// if there is, check by synchronization fields and replace
		} else {
			filter := bson.M{"_id": document["_id"], config.SyncDateField: bson.M{"$ne": document[config.SyncDateField]}}
			res, err := collectionDst.ReplaceOne(context.Background(), filter, document)
			if res.ModifiedCount != 0 {
				logger.Println("replace", document)
			}
			if err != nil {
				logger.Fatal(err)
			}
		}
	}
}
