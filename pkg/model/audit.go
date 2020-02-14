package model

import "time"

// FieldAuditRecord is the record of a single change to a specific field of a record.
type FieldAuditRecord struct {
	FieldID    string      `json:"fieldId"`
	FieldValue interface{} `json:"fieldValue"`
	UpdatedBy  string      `json:"updatedBy"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}
