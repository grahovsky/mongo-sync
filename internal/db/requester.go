package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mongo-sync/internal/config"

	"go.mongodb.org/mongo-driver/bson"
)

func RequestNewDocs(docs chan<- Doc, requestTime time.Time) error {
	clientSrc, _ := NewClient(config.Settings.DBSource)
	dbSrc := clientSrc.Database(config.Settings.DBSource.Name)

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

		found := 0
		for cur.Next(context.Background()) {
			var document bson.M
			if err := cur.Decode(&document); err != nil {
				slog.Error(err.Error())
			}
			docs <- Doc{document: document, collection: collectionName}
			slog.Debug(fmt.Sprintf("found %s", document))
			found++
		}
		if found > 0 {
			slog.Info(fmt.Sprintf("found total: %d", found))
		}

		if err := cur.Err(); err != nil {
			slog.Error(err.Error())
		}
		cur.Close(context.Background())
	}

	close(docs)
	return nil
}
