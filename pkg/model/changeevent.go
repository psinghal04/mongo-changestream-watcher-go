package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChangeEvent is the customized representation of a MongoDB change stream event that is captured and processed by
// this application.
type ChangeEvent struct {
	ID                ResumeToken         `bson:"_id" json:"_id"`
	User              string              `bson:"user" json:"user"`
	Timestamp         primitive.Timestamp `bson:"timestamp" json:"timestamp"`
	OperationType     string              `bson:"operationType" json:"operationType"`
	Database          string              `bson:"database" json:"database"`
	Collection        string              `bson:"collection" json:"collection"`
	DocumentKey       primitive.ObjectID  `bson:"documentKey" json:"documentKey"`
	FullDocument      primitive.D         `bson:"fullDocument" json:"fullDocument"`
	UpdateDescription struct {
		UpdatedFields map[string]interface{} `bson:"updatedFields" json:"updatedFields"`
		RemovedFields interface{}            `bson:"removedFields" json:"removedFields"`
	} `bson:"updateDescription" json:"updateDescription"`
}

// ResumeToken denotes the token associated with a MongoDB change stream event, which may be used to resume receiving change stream events from
// a point in the past.
type ResumeToken struct {
	TokenData interface{} `bson:"_data" json:"_data"`
}
