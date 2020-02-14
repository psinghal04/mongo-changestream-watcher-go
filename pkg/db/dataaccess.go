package db

import (
	"context"
	"strconv"

	"mongo-changestreams/pkg/config"
	"mongo-changestreams/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DataAccess provides data access functions for the underlying MongoDB collections - both the audited collection, as well
// as the collection containing audit records.
type DataAccess struct {
	DbClient *mongo.Client
}

// InitializeDataAccess initializes a connection to the indicated MongoDB database.
func InitializeDataAccess(dbURL string) (*DataAccess, error) {
	dao := DataAccess{}

	c, err := mongo.NewClient(options.Client().ApplyURI(dbURL))
	if err != nil {
		return &dao, err
	}

	err = c.Connect(context.Background())
	if err != nil {
		return &dao, err
	}

	dao.DbClient = c

	return &dao, nil
}

// ChangeLogger saves the supplied change event to a persistent store.
type ChangeLogger interface {
	SaveChangeEvent(ce model.ChangeEvent) error
}

// ChangeLogTracker fetches metadata of stored change events.
type ChangeLogTracker interface {
	GetResumeToken() (model.ResumeToken, error)
}

// MongoDBChangeLogTracker fetches metadata of change events stored in the MongoDB audit collection.
type MongoDBChangeLogTracker struct {
	Config config.Configuration
	Dao    *DataAccess
}

// GetResumeToken returns the mongo stream token for the last change stream event that was recorded in the audit database.
// This may be used to resume receiving change events from the point of the last change event.
func (m *MongoDBChangeLogTracker) GetResumeToken() (model.ResumeToken, error) {
	coll := m.Dao.DbClient.Database(m.Config.AuditDatabase).Collection(m.Config.AuditCollection)

	var opts options.FindOptions
	var l = int64(1)
	opts.Limit = &l
	opts.Sort = map[string]int{"timestamp": -1}
	opts.Projection = map[string]int{"_id": 1}
	ctx := context.Background()
	cur, err := coll.Find(ctx, bson.D{}, &opts)
	if err != nil {
		return model.ResumeToken{}, err
	}
	defer cur.Close(ctx)

	var ce model.ChangeEvent
	if cur.Next(ctx) {
		raw := cur.Current
		err := bson.Unmarshal(raw, &ce)
		if err != nil {
			return model.ResumeToken{}, err
		}
	}

	return ce.ID, nil
}

// MongoDBChangeLogger can be used to save a change event to a MongoDB collection designated to store
// audit records.
type MongoDBChangeLogger struct {
	Config config.Configuration
	Dao    *DataAccess
}

// SaveChangeEvent saves the change event to the audit database.
func (m *MongoDBChangeLogger) SaveChangeEvent(ce model.ChangeEvent) error {
	ctx := context.Background()
	coll := m.Dao.DbClient.Database(m.Config.AuditDatabase).Collection(m.Config.AuditCollection)

	update := bson.M{
		"$set": ce,
	}

	_, err := coll.UpdateOne(ctx, bson.D{{Key: "_id", Value: ce.ID}}, update, options.Update().SetUpsert(true))
	return err
}

// TraverseForFieldValue walks through the supplied document of type primitive.D to locate and fetch the value of the
// a specific field, whose field path is supplied as an array (for example, {"arr", "0", "field1"} denotes
// the field "field1" of the first element of the array "arr").
func TraverseForFieldValue(f []string, payload primitive.D) interface{} {
	f1 := f[0]
	v1 := payload.Map()[f1]
	if len(f) == 1 {
		return v1
	}

	f2 := f[1]
	if i, err := strconv.ParseInt(f2, 10, 64); err == nil {
		arr := v1.(primitive.A)
		v1 = arr[i]
	}

	if len(f) == 2 {
		return v1
	}

	return TraverseForFieldValue(f[2:], v1.(primitive.D))
}
