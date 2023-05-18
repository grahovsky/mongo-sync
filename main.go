package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/ini.v1"
)

type Config struct {
	SyncDateField string
	FirstSyncDate string
	LogFile       string
	Timeout       int
	NumWorkers    int
	SrcConnStr    string
	SrcDbName     string
	SrcUsername   string
	SrcPassword   string
	SrcAuthsorce  string
	DstConnStr    string
	DstDbName     string
	DstUsername   string
	DstPassword   string
	DstAuthsorce  string
}

type Conn struct {
	ConnString string
	Username   string
	Password   string
	AuthSourc  string
}

var config *Config

type Doc struct {
	document   bson.M
	collection string
}

func loadConfig(filename string) error {
	cfg, err := ini.Load(filename)
	if err != nil {
		return err
	}

	sectionSrc := cfg.Section("source")
	sectionDst := cfg.Section("destination")
	sectionCmn := cfg.Section("common")

	config = &Config{
		SyncDateField: sectionCmn.Key("sync_date_field").String(),
		FirstSyncDate: sectionCmn.Key("first_sync_date").String(),
		LogFile:       sectionCmn.Key("log_file").String(),
		SrcConnStr:    sectionSrc.Key("src_conn_str").String(),
		SrcDbName:     sectionSrc.Key("src_db_name").String(),
		SrcUsername:   sectionSrc.Key("src_username").String(),
		SrcPassword:   sectionSrc.Key("src_password").String(),
		SrcAuthsorce:  sectionSrc.Key("src_auth_sorce").String(),
		DstConnStr:    sectionDst.Key("dst_conn_str").String(),
		DstDbName:     sectionDst.Key("dst_db_name").String(),
		DstUsername:   sectionDst.Key("dst_username").String(),
		DstPassword:   sectionDst.Key("dst_password").String(),
		DstAuthsorce:  sectionDst.Key("dst_auth_sorce").String(),
	}

	timeout, err := sectionCmn.Key("timeout").Int()
	if err != nil {
		log.Fatal(err)
	}
	config.Timeout = timeout

	numWorkers, err := sectionCmn.Key("num_workers").Int()
	if err != nil {
		log.Fatal(err)
	}
	config.NumWorkers = numWorkers

	return nil
}

func GetClient(c Conn) *mongo.Client {
	clientOptions := options.Client().ApplyURI(c.ConnString).SetAuth(options.Credential{
		Username:   c.Username,
		Password:   c.Password,
		AuthSource: c.AuthSourc,
	})

	clientDst, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return clientDst
}

func worker(jobs <-chan Doc, wg *sync.WaitGroup) {
	defer wg.Done()

	// Connecting to a Destination Base
	conn := Conn{
		config.DstConnStr,
		config.DstUsername,
		config.DstPassword,
		config.DstAuthsorce,
	}
	clientDst := GetClient(conn)
	defer clientDst.Disconnect(context.Background())

	dbDst := clientDst.Database(config.DstDbName)

	for doc := range jobs {

		document := doc.document
		collectionDst := dbDst.Collection(doc.collection)

		// look for documents in the destination database by _id
		count, err := collectionDst.CountDocuments(context.Background(), bson.M{"_id": document["_id"]})
		if err != nil {
			log.Fatal(err)
		}

		// if there is no document, add them
		if count == 0 {
			log.Println("insert", document)
			_, err := collectionDst.InsertOne(context.Background(), document)
			if err != nil {
				log.Fatal(err)
			}
			// if there is, check by synchronization fields and replace
		} else {
			filter := bson.M{"_id": document["_id"], "updated_at": bson.M{"$ne": document["updated_at"]}}
			res, err := collectionDst.ReplaceOne(context.Background(), filter, document)
			if res.ModifiedCount != 0 {
				log.Println("replace", document)
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	// Read config file
	err := loadConfig("config")
	if err != nil {
		log.Fatal(err)
	}

	// Create a log file
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// Set log output to file
	log.SetOutput(logFile)

	// Connect to source database
	conn := Conn{
		config.SrcConnStr,
		config.SrcUsername,
		config.SrcPassword,
		config.SrcAuthsorce,
	}
	clientSrc := GetClient(conn)
	defer clientSrc.Disconnect(context.Background())
	dbSrc := clientSrc.Database(config.SrcDbName)

	// Set the start time of the check synchronization
	var checkTime time.Time
	switch config.FirstSyncDate {
	case "":
		checkTime = time.Now().Add(-time.Hour * 2)
	default:
		checkTime, err = time.Parse("2006-01-02", config.FirstSyncDate)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println(checkTime.String())
	timeout := time.Second * time.Duration(config.Timeout)

	for {
		log.Println("Start sync iteration")

		// Getting a list of collections in the source base
		collectionNames, err := dbSrc.ListCollectionNames(context.Background(), bson.M{})
		if err != nil {
			log.Fatal(err)
		}

		// time to filter data for processing
		requestTime := checkTime.Add(-2 * timeout)

		// Update the check time to the beginning of the request
		checkTime = time.Now()

		// create channel
		jobs := make(chan Doc, 500)

		var wg sync.WaitGroup
		wg.Add(config.NumWorkers)

		// create workers
		for w := 1; w <= config.NumWorkers; w++ {
			go worker(jobs, &wg)
		}

		// fetch new documents from all collections of the source base
		for _, collectionName := range collectionNames {
			collectionSrc := dbSrc.Collection(collectionName)

			filter := bson.M{"updated_at": bson.M{"$gt": requestTime}}
			cur, err := collectionSrc.Find(context.Background(), filter)
			if err != nil {
				log.Fatal(err)
			}
			defer cur.Close(context.Background())

			for cur.Next(context.Background()) {

				var document bson.M
				if err := cur.Decode(&document); err != nil {
					log.Fatal(err)
				}
				jobs <- Doc{document: document, collection: collectionName}

				log.Println("finded", document)
			}

			if err := cur.Err(); err != nil {
				log.Fatal(err)
			}
		}

		close(jobs)
		wg.Wait()

		log.Println("End sync iteration")

		duration := time.Since(checkTime)
		if duration < timeout {
			time.Sleep(timeout - duration)
		}

	}
}
