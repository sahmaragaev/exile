package db

import "go.mongodb.org/mongo-driver/mongo"

var (
	UserCollection   *mongo.Collection
	ThreadCollection *mongo.Collection
)

func InitializeCollections() {
	UserCollection = getDbCollection("users")
	ThreadCollection = getDbCollection("threads")
}