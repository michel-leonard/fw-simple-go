package config

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
)

type FileConfig struct {
	Accept []string `json:"accept"`
	Reject []string `json:"reject"`
}

// Filled accordingly with a JSON file
type Config struct {
	Files         map[string]FileConfig       `json:"files"`
	AcceptRegexps map[string][]*regexp.Regexp // Filled by the software
	RejectRegexps map[string][]*regexp.Regexp // Filled by the software
	Name          string                      `json:"firewall-name"`
	Path          string                      `json:"path-iptables-ipset"`
	RejectBitLen  int                         `json:"reject-bitmask-length"`
	RejectTimeout int                         `json:"reject-timeout"`
}

// Unmarshal the data from the JSON file
func ReadConfig(filename string) Config {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file %v", err)
	}
	log.Printf("Configuration file '%s' unmarshalled.\n", filename)
	return config
}
