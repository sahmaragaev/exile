package db

import (
	"context"
	"errors"
	"exile-telegram-bot/pkg/models"
	"exile-telegram-bot/pkg/utils"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func Connect(connectionString string) {
	var err error

	Client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	err = Client.Ping(context.TODO(), nil)
	log.Println("Connected to MongoDB!")
}

func getDbCollection(collectionName string) *mongo.Collection {
	return Client.Database("apex").Collection(collectionName)
}

func EnsureUserExists(telegramId string) (primitive.ObjectID, error) {
	filter := bson.M{"username": telegramId}
	var user models.User

	err := UserCollection.FindOne(context.Background(), filter).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return createUser(telegramId)
	} else if err != nil {
		log.Printf("Error finding user in MongoDB: %v", err)
		return primitive.NilObjectID, err
	}

	return user.ID, nil
}

func createUser(telegramId string) (primitive.ObjectID, error) {
	password := utils.GenerateRandomPassword(10)
	user := models.User{
		ID:          primitive.NewObjectID(),
		DisplayName: telegramId,
		Username:    telegramId,
		Password:    password,
	}

	result, err := UserCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Printf("Error inserting new user into MongoDB: %v", err)
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}
