package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Otp struct {
    ID              primitive.ObjectID `bson:"_id,omitempty"`
    Otp             string             `bson:"otp"`
    CreationDate    time.Time          `bson:"creationDate"`
    ModificationDate time.Time          `bson:"modificationDate"`
    Status          string             `bson:"status"`
    ValidAttempts   int                `bson:"validAttempts"`
    RetryAttempts   int                `bson:"retryAttempts"`
    UserPhoneNumber string             `bson:"userPhoneNumber"`
    UserID          string             `bson:"userId"`
}
