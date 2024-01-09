package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Doc struct {
	document   bson.M
	collection string
}
