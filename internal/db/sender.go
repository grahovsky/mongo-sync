package db

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"mongo-sync/internal/config"

	"go.mongodb.org/mongo-driver/bson"
)

func Worker(jobs <-chan Doc, wg *sync.WaitGroup) error {
	defer wg.Done()

	clientDst, _ := NewClient(config.Settings.DBDestination)
	dbDst := clientDst.Database(config.Settings.DBDestination.Name)

	for doc := range jobs {

		document := doc.document
		collectionDst := dbDst.Collection(doc.collection)

		// look for documents in the destination database by _id
		count, err := collectionDst.CountDocuments(context.Background(), bson.M{"_id": document["_id"]})
		if err != nil {
			return err
		}

		// if there is no document, add them
		if count == 0 {
			slog.Info(fmt.Sprintf("insert %s", document))
			_, err := collectionDst.InsertOne(context.Background(), document)
			if err != nil {
				slog.Error(err.Error())
			}
			// if there is, check by synchronization fields and replace
		} else {
			filter := bson.M{"_id": document["_id"], config.Settings.Common.SyncDateField: bson.M{"$ne": document[config.Settings.Common.SyncDateField]}}
			res, err := collectionDst.ReplaceOne(context.Background(), filter, document)
			if res.ModifiedCount != 0 {
				slog.Info(fmt.Sprintf("replace %s", document))
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}
