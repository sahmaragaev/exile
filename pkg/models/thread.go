package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Thread struct {
    ID          primitive.ObjectID `bson:"_id"`
    UserID      string             `bson:"userId"`
    ThreadID    string             `bson:"threadId"`
    AssistantID string             `bson:"assistantId"`
}