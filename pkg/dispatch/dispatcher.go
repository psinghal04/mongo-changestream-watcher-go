package dispatch

import (
	"log"

	"mongo-changestreams/pkg/config"
	"mongo-changestreams/pkg/db"
	"mongo-changestreams/pkg/model"
)

// ChangeEventDispatcherFunc is a type of function that can act on a change event.
type ChangeEventDispatcherFunc func(ce model.ChangeEvent) error

// GetSaveChangeEventFunc returns a function that can save a change event to an audit store.
func GetSaveChangeEventFunc(c config.Configuration, dao *db.DataAccess) ChangeEventDispatcherFunc {
	lt := db.MongoDBChangeLogger{Config: c, Dao: dao}

	return func(ce model.ChangeEvent) error {
		log.Printf("Saving change event of type %s for collection %s for database %s for record %v\n", ce.OperationType, ce.Collection, ce.Database, ce.DocumentKey)
		err := lt.SaveChangeEvent(ce)
		if err != nil {
			log.Printf("ERROR: failed to save change event ID %v to audit DB due to error: %v\n", ce.ID.TokenData, err)
			return err
		}

		return nil
	}
}
