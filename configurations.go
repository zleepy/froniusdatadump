package main

import (
	"encoding/json"
	"io/ioutil"
)

const configFileName = "config.json"

type configurations struct {
	Source         source
	Sink           sink
	VerboseLogging bool
}

type source struct {
	APIUri     string
	SleepInSec uint16
}

type sink struct {
	APIUri   string
	Database string
}

func readConfigFile(config *configurations) error {
	buff, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buff, config)
	return err
}
