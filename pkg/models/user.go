package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id"`
	Username    string             `bson:"username"`
	DisplayName string             `bson:"displayName"`
	Password    string             `bson:"password"`
}
