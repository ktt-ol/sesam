package main

import (
	"fmt"
	"github.com/ktt-ol/sesam/internal/conf"
	"github.com/ktt-ol/sesam/internal/mqtt"
	"github.com/ktt-ol/sesam/internal/web"
	"github.com/ktt-ol/sesam/internal/wikiauth"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	config := conf.LoadConfig("config.toml")
	setupLogging(config.Logging)

	logrus.WithFields(logrus.Fields{
		"mqttUrl":      config.Mqtt.Url,
		"mqttUser": config.Mqtt.Username,
		"serving":  fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		"https":    config.Server.Https,
	}).Info("Sesam is starting...")

	//auth := wikiauth.NewLocalFilesAuth(&config.AuthLocal)
	auth := wikiauth.NewOnlineAuth(&config.AuthOnline)

	//mqtt.EnableMqttDebugLogging()
	mqttHandler := mqtt.NewMqttHandler(config.Mqtt)

	web.StartWeb(config.Server, auth, mqttHandler)
}

type StdErrLogHook struct {
}

func (h *StdErrLogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
func (h *StdErrLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	fmt.Fprintf(os.Stderr, line)
	return nil
}

func setupLogging(config conf.LoggingConf) {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	if config.DebugLogging {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if config.Logfile == "" {
		logrus.SetOutput(os.Stdout)
	} else {
		// https://github.com/sirupsen/logrus/issues/227
		file, err := os.OpenFile(config.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.Warnf("Failed to log to file '%s', using default stderr.", config.Logfile)
		}
		logrus.AddHook(&StdErrLogHook{})
	}
}
