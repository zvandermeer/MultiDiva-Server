package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var CurrentConfigVersion int = 2

type ConfigData struct {
	ConfigVersion      int    `yaml:"config_version"`
	Port               string `yaml:"bind_port"`
	MaxConcurrentUsers int    `yaml:"max_concurrent_users"`
	MaxRoomSize        int    `yaml:"max_room_size"`
	MaxRoomCount       int    `yaml:"max_rooms"`
	BindAddress        string `yaml:"bind_address"`
}

func NewConfigData() (config ConfigData) {
	config.ConfigVersion = CurrentConfigVersion
	config.Port = "9988"
	config.MaxConcurrentUsers = 100
	config.MaxRoomSize = 6
	config.MaxRoomCount = 15
	config.BindAddress = "0.0.0.0"
	return
}

var ConfigLocation string = "./MultiDiva_Server_Config.yml"

func LoadConfig() (cfg ConfigData) {
	if _, err := os.Stat(ConfigLocation); os.IsNotExist(err) {
		writeConfig(NewConfigData())
	}

	cfg = readConfig()

	if cfg.ConfigVersion < CurrentConfigVersion {
		cfg.ConfigVersion = CurrentConfigVersion
		writeConfig(cfg)
	}

	return
}

func readConfig() (myConfig ConfigData) {
	myConfig = NewConfigData()

	dat, err := os.ReadFile(ConfigLocation)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(dat, &myConfig)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func writeConfig(data ConfigData) {
	yamlOutput, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create(ConfigLocation)
	if err != nil {
		fmt.Println(err)
	}

	_, err = f.Write(yamlOutput)
	if err != nil {
		return
	}
}
