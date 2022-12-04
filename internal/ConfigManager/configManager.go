package ConfigManager

import (
	"fmt"
	"os"

	"github.com/ovandermeer/MultiDiva-Server/internal/DataTypes"

	"gopkg.in/yaml.v3"
)

var ConfigLocation string = "./MultiDiva_Server_Config.yml"

func LoadConfig() (cfg DataTypes.ConfigData) {
	if _, err := os.Stat(ConfigLocation); os.IsNotExist(err) {
		writeConfig(DataTypes.NewConfigData())
	}

	cfg = readConfig()

	if cfg.ConfigVersion < DataTypes.CurrentConfigVersion {
		cfg.ConfigVersion = DataTypes.CurrentConfigVersion
		writeConfig(cfg)
	}

	return
}

func readConfig() (myConfig DataTypes.ConfigData) {
	myConfig = DataTypes.NewConfigData()

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

func writeConfig(data DataTypes.ConfigData) {
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
