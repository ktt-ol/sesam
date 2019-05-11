package conf

import (
	"github.com/BurntSushi/toml"
	"log"
)

func LoadConfig(configFile string) TomlConfig {
	config := &TomlConfig{}
	if _, err := toml.DecodeFile(configFile, config); err != nil {
		log.Fatal("Could not read config file.", err)
	}

	return *config
}

type TomlConfig struct {
	Logging    LoggingConf
	Server     ServerConf
	Mqtt       MqttConf
	AuthLocal  AuthLocal
	AuthOnline AuthOnline
}

type LoggingConf struct {
	DebugLogging bool
	Logfile      string
}

type ServerConf struct {
	Host        string
	Port        int
	Https       bool
	CertKeyFile string
	CertFile    string
	KeysFile    string
}

type MqttConf struct {
	Url      string
	Username string
	Password string
	// if empty, the system certificates are used
	CertFile                  string
	StatusTopic               string
	MainDoorBuzzerTopic       string
	GlassDoorBuzzerTopic      string
	DoorDownstairsBuzzerTopic string
}

type AuthLocal struct {
	UserDirectory string
	GroupPageFile string
}

type AuthOnline struct {
	WikiBaseUrl string
	AuthToken   string
}
