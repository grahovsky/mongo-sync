package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mongo-sync/internal/config"

	"go.mongodb.org/mongo-driver/bson"
)

func RequestNewDocs(jobs chan<- Doc, requestTime time.Time) error {
	clientSrc, _ := NewClient(config.Settings.DBSource)
	dbSrc := clientSrc.Database(config.Settings.DBSource.Name)

	slog.Info("Start sync iteration")

	// Getting a list of collections in the source base
	collectionNames, err := dbSrc.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	// fetch new documents from all collections of the source base
	for _, collectionName := range collectionNames {
		collectionSrc := dbSrc.Collection(collectionName)

		filter := bson.M{config.Settings.Common.SyncDateField: bson.M{"$gt": requestTime}}
		cur, err := collectionSrc.Find(context.Background(), filter)
		if err != nil {
			return err
		}

		for cur.Next(context.Background()) {
			var document bson.M
			if err := cur.Decode(&document); err != nil {
				slog.Error(err.Error())
			}
			jobs <- Doc{document: document, collection: collectionName}

			slog.Info(fmt.Sprintf("finded %s", document))
		}

		if err := cur.Err(); err != nil {
			slog.Error(err.Error())
		}
		cur.Close(context.Background())
	}

	return nil
}
