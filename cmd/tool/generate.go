package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"mongo-sync/internal/config"
	"mongo-sync/internal/db"
	"mongo-sync/internal/logger"
)

type Doc struct {
	Text      string
	UpdatedAt time.Time `bson:"updated_at"`
}

func main() {
	config.Init()
	logger.Set(logger.Options{Level: config.Settings.Log.Level, AddSource: false})

	records := make([]interface{}, 0)

	for i := 0; i < 5000; i++ {
		rec := Doc{UpdatedAt: time.Now(), Text: fmt.Sprintf("test Item%d", i)}
		records = append(records, rec)
	}
	fmt.Println(config.Settings.DBSource)
	err := db.RecordMany(context.Background(),
		&db.SaveReq{Collection: "testCollection1", Data: records},
		config.Settings.DBSource)
	if err != nil {
		slog.Error(err.Error())
	}
}
