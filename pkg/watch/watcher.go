package watch

import (
	"context"
	"log"
	"strings"

	"mongo-changestreams/pkg/config"
	"mongo-changestreams/pkg/db"
	"mongo-changestreams/pkg/dispatch"
	"mongo-changestreams/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChangeEventWatcher watches a change stream for change events, and dispatches received changes events
// via one or more change event dispatch functions.
type ChangeEventWatcher interface {
	// WatchChangeEvents resumes watching the change events from the supplied change event token, passing any received
	// event to the supplied dispatch functions for handling.
	WatchChangeEvents(resumeToken *model.ResumeToken, disps ...dispatch.ChangeEventDispatcherFunc) error
}

// MongoDBChangeStreamWatcher watches a MongoDB change stream for change events and reacts to those events.
type MongoDBChangeStreamWatcher struct {
	Config config.Configuration
	Dao    *db.DataAccess
}

// WatchChangeEvents starts watching the mongo change stream for the MongoDB collection associated with the  MongoDBChangeStreamWatcher. If a valid
// resume token is provided, the stream starts from that point.
func (m *MongoDBChangeStreamWatcher) WatchChangeEvents(resumeToken model.ResumeToken, dispatchFuncs ...dispatch.ChangeEventDispatcherFunc) error {
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	if resumeToken.TokenData != nil {
		opts.SetResumeAfter(resumeToken)
	}

	coll := m.Dao.DbClient.Database(m.Config.AppDatabase).Collection(m.Config.AppCollection)
	ctx := context.Background()
	watchCursor, err := coll.Watch(ctx, buildChangeStreamAggregationPipeline(), opts)
	if err != nil {
		log.Printf("ERROR: encountered error watching for change stream data, error is %v\n", err)
		return err
	}
	defer watchCursor.Close(ctx)

	// wait for the next change stream data to become available
	log.Println("Mongo stream watcher launched, waiting for change events...")
	for watchCursor.Next(ctx) {
		ce, err := m.extractChangeEvent(watchCursor.Current)
		if err != nil {
			log.Printf("ERROR: failed to extract change event from change stream data, error is %v\n ", err)
			return err
		}

		for _, dispFunc := range dispatchFuncs {
			go func(f func(ev model.ChangeEvent) error, e model.ChangeEvent) {
				log.Printf("Received change event %s, dispatching to audit logger\n", e.ID)
				err := f(e)
				if err != nil {
					log.Printf("ERROR: failed to invoke change event dispatch function, err is : %v\n", err)
				}
			}(dispFunc, ce)
		}
	}

	return nil
}

// extractChangeEvent transforms the raw data received from the MongoDB change stream to the model.ChangeEvent type.
func (m *MongoDBChangeStreamWatcher) extractChangeEvent(rawChange bson.Raw) (model.ChangeEvent, error) {
	var ce model.ChangeEvent
	err := bson.Unmarshal(rawChange, &ce)
	if err != nil {
		return ce, err
	}

	// extract the user name of the change owner, assuming this is contained within the configured field of the full document.
	// In case the field is not present, or if there is an error, this defaults to a blank value.
	ce.User, _ = db.TraverseForFieldValue(strings.Split(m.Config.UserFieldPath, "."), ce.FullDocument).(string)

	// if there is no need to record the full document for the current operation type, remove it
	if !m.Config.CaptureFullDocument[ce.OperationType] {
		ce.FullDocument = nil
	}

	return ce, nil
}

// buildChangeStreamAggregationPipeline builds a MongoDB aggregation pipeline to reshape the change stream data received from MongoDB in
// the format of our change events. See model.ChangeEvent.
func buildChangeStreamAggregationPipeline() mongo.Pipeline {
	pipeline := mongo.Pipeline{bson.D{{Key: "$addFields", Value: bson.D{{Key: "timestamp", Value: "$clusterTime"}, {Key: "database", Value: "$ns.db"}, {Key: "collection", Value: "$ns.coll"}, {Key: "documentKey", Value: "$documentKey._id"}}}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "timestamp", Value: 1}, {Key: "operationType", Value: 1}, {Key: "database", Value: 1}, {Key: "collection", Value: 1}, {Key: "documentKey", Value: 1}, {Key: "fullDocument", Value: 1}, {Key: "updateDescription", Value: 1}}}}}

	return pipeline
}
