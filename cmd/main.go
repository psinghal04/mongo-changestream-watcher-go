package main

import (
	"log"
	"os"

	"mongo-changestreams/pkg/config"
	"mongo-changestreams/pkg/db"
	"mongo-changestreams/pkg/dispatch"
	"mongo-changestreams/pkg/watch"
)

func main() {
	c := config.GetConfiguration()

	auditDao, err := db.InitializeDataAccess(c.AuditDBUrl)
	if err != nil {
		log.Printf("ERROR, failed to initialize data access to the Audit database due to error: %v\n", err)
		os.Exit(1)
	}
	appDao, err := db.InitializeDataAccess(c.AppDBUrl)
	if err != nil {
		log.Printf("ERROR, failed to initialize data access to the App database due to error: %v\n", err)
		os.Exit(1)
	}

	r := db.MongoDBChangeLogTracker{Config: c, Dao: auditDao}
	resToken, err := r.GetResumeToken()
	if err != nil {
		log.Printf("ERROR, failed to look up resume token with error: %v\n", err)
		os.Exit(1)
	}

	w := watch.MongoDBChangeStreamWatcher{Config: c, Dao: appDao}
	err = w.WatchChangeEvents(resToken, dispatch.GetSaveChangeEventFunc(c, auditDao))
	if err != nil {
		log.Printf("ERROR, failed to watch change stream due to error: %v\n", err)
		os.Exit(1)
	}
}
