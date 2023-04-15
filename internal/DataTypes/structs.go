package dataTypes

var CurrentConfigVersion int = 2

type ConfigData struct {
	ConfigVersion      int    `yaml:"config_version"`
	Port               string `yaml:"default_bind_port"`
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
