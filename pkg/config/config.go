package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Configuration contains configurable parameters for the application. These may be driven by the configuration file config.json, or by environment variables.
type Configuration struct {
	AppDBUrl      string `json:"appDbUrl"`
	AppDatabase   string `json:"appDatabaseName"`
	AppCollection string `json:"appDatabaseCollection"`
	UserFieldPath string `json:"userFieldPath"`

	AuditDBUrl      string `json:"auditDbUrl"`
	AuditDatabase   string `json:"auditDatabaseName"`
	AuditCollection string `json:"auditDatabaseCollection"`

	CaptureFullDocument map[string]bool `json:"fullDocRecordOperations"`
	Version             string          `json:"version"`
}

// GetConfiguration loads and fetches the application configuration.
func GetConfiguration() Configuration {
	var cf string
	if v, ok := os.LookupEnv("CONFIG_FILE"); ok {
		cf = filepath.FromSlash(v)
	} else {
		cf = filepath.FromSlash("./config.json")
	}

	c := loadConfigurationFromFile(cf)

	//override with values from environment variables, if present
	if v, ok := os.LookupEnv("APP_DB_URL"); ok {
		c.AppDBUrl = v
	}
	if v, ok := os.LookupEnv("APP_DB_NAME"); ok {
		c.AppDatabase = v
	}
	if v, ok := os.LookupEnv("APP_COLLECTION"); ok {
		c.AppCollection = v
	}
	if v, ok := os.LookupEnv("AUDIT_DB_URL"); ok {
		c.AuditDBUrl = v
	}
	if v, ok := os.LookupEnv("AUDIT_DB_NAME"); ok {
		c.AuditDatabase = v
	}
	if v, ok := os.LookupEnv("AUDIT_COLLECTION"); ok {
		c.AuditCollection = v
	}
	if v, ok := os.LookupEnv("API_VERSION"); ok {
		c.Version = v
	}

	return c
}

func loadConfigurationFromFile(filePath string) Configuration {
	c := Configuration{}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("WARNING, config file %s does not exist, will use default config, or config values from environment variables", filePath)
		return c
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("WARNING, failed to open config file %s, will use default config, or config values from environment variables", filePath)
		return c
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		log.Printf("WARNING, failed to decode configuration values from config file %s, will use default config, or config values from environment variables", filePath)
		return c
	}

	return c
}
