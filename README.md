# MongoDB Change Stream Watcher
This program is capable of watching the change stream for a given MongoDB collection, recording the change stream
events as audit records in a designated audit collection. If the watcher gets disconnected from the change stream 
for any reason, it can resume watching the stream from the point it left off before being disconnected, ensuring no 
change events are lost.

## Configuration
The watcher program supports several configurable parameters, as illustrated below
```json
{
  "appDbUrl": "mongodb://@localhost:27017",
  "appDatabaseName": "test",
  "appDatabaseCollection": "streamtest",
  "userFieldPath": "updatedBy",
  "auditDbUrl": "mongodb://@localhost:27017",
  "auditDatabaseName": "test",
  "auditDatabaseCollection": "audit",
  "fullDocRecordOperations": {
    "insert": true
  },
  "version": "1.0"
}
```
**appDBUrl**: URL for the MongoDB instance that contains the collection be watched. <br>  
**appDatabaseName**: Name of the MongoDB database that contains the collection be watched. <br>  
**appDatabaseCollection**: Name of the MongoDB collection to be watched. <br>  
**userFieldPath**: The fully qualified path of the field in the watched collection that contains the user ID of 
the user who made the change. This assumes that the application writing to the collection tracks and stores the user ID in this field. 
For example, if the watched collection's structure is like this:
```json
{
  _id: "123";
  amount: 123.34;
  changeDetails: 
     {
       changedBy: "abc",
       changedOn: "2020-02-14T15:03:27Z"
     }
}
```
In this case, the value of the userFieldPath field would be "changeDetails.changedBy" <br>  
**auditDbUrl**: URL for the MongoDB instance that contains the collection where the audit records are to be stored. <br>  
**auditDatabaseName**: Name of the MongoDB database where the audit records are to be stored. <br>  
**auditDatabaseCollection**: Name of the MongoDB collection for storing audit records.<br>  
**fullDocRecordOperations**: Determines the change event types for which the entire document will be stored in the audit record. In some situations, you might want to capture the
current state of the entire document (from the collection being watched) for all types of changes (insert/update/delete, etc.). 

### Structure of Audit Records
The following is an example of an audit record that is created by the watcher:
```json
{ 
    "_id" : {
        "_data" : "825E446661000000012B022C0100296E5A1004EC1E76078DCE4C489A2BFE17218EC79F46645F696400645C5D85C62FEF357A165CCABF0004"
    }, 
    "collection" : "streamtest", 
    "database" : "test", 
    "documentKey" : "5c5d85c62fef357a165ccabf", 
    "fullDocument" : null, 
    "operationType" : "update", 
    "timestamp" : "2020-02-14T15:03:27Z", 
    "updateDescription" : {
        "updatedFields" : {
            "lineItems.0.procedures.0.procedureModCodes.0" : "332"
        }, 
        "removedFields" : [

        ]
    }, 
    "user" : "tcadmin"
}
```
Whether or not the "fullDocument" field contains the state of the entire document at the time of the change event depends on the "fullDocRecordOperations"
configuration described earlier. The audit records use the change stream token for their unique ID. When the watcher comes
online, it fetches the most recent change token to determine the point from which to watch the change stream. The updateDescription section
contains details of the fields that were modified in the affected document of the watched collection as part of the change
event.