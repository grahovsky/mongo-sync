package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mongo-sync/internal/config"
)

func RecordOne(record *SaveReq, conn config.DbConn) error {
	ctx := context.Background()

	// Connect to source database
	clientSrc, err := NewClient(conn)
	if err != nil {
		return err
	}

	dbDst := clientSrc.Database(config.Settings.DBDestination.Name)
	slog.Debug(fmt.Sprintf("start saving with mongo db %s", config.Settings.DBDestination.Name))

	collectionSrc := dbDst.Collection(record.Collection)
	res, err := collectionSrc.InsertOne(ctx, record.Data)
	if err != nil {
		return err
	}
	slog.Debug(fmt.Sprintf("saved %s", res))

	return nil
}

func RecordMany(ctx context.Context, records *SaveReq, conn config.DbConn) error {
	// Connect to source database
	clientSrc, err := NewClient(conn)
	if err != nil {
		return err
	}

	dbDst := clientSrc.Database(config.Settings.DBDestination.Name)
	slog.Debug(fmt.Sprintf("start saving with mongo db %s", conn.Name))

	collectionSrc := dbDst.Collection(records.Collection)
	res, err := collectionSrc.InsertMany(ctx, records.Data)
	if err != nil {
		return err
	}
	slog.Debug(fmt.Sprintf("saved %s", res))

	return nil
}

func CleanClients() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, conn := range Pull.Clients {
		errDisc := conn.Disconnect(ctx)
		if errDisc != nil {
			slog.Error(errDisc.Error())
		}
	}
}
