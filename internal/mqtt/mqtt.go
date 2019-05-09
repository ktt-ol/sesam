package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/sesam/internal/conf"
	"github.com/sirupsen/logrus"
)

const CLIENT_ID = "sesam"

// the amount of ms the door will buzz
const BUZZER_DURATION = 4004

var mqttLogger = logrus.WithField("where", "mqtt")

type Door int8

const DoorOuter = Door(0)
const DoorInnerGlass = Door(1)
const DoorInnerMetal = Door(2)

type MqttHandler struct {
	client mqtt.Client
	status string
	conf   conf.MqttConf
}

func EnableMqttDebugLogging() {
	stdLogWriter := logrus.StandardLogger().Writer()
	mqtt.ERROR.SetOutput(stdLogWriter)
	mqtt.CRITICAL.SetOutput(stdLogWriter)
	mqtt.WARN.SetOutput(stdLogWriter)
	mqtt.DEBUG.SetOutput(stdLogWriter)
}

func NewMqttHandler(conf conf.MqttConf) *MqttHandler {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(conf.Url)

	if conf.Username != "" {
		opts.SetUsername(conf.Username)
	}
	if conf.Password != "" {
		opts.SetPassword(conf.Password)
	}

	certs := defaultCertPool(conf.CertFile)
	tlsConf := &tls.Config{
		RootCAs: certs,
	}
	opts.SetTLSConfig(tlsConf)

	opts.SetClientID(CLIENT_ID)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	handler := MqttHandler{status: conf.StatusTopic, conf: conf}
	opts.SetOnConnectHandler(handler.onConnect)
	opts.SetConnectionLostHandler(handler.onConnectionLost)

	handler.client = mqtt.NewClient(opts)
	if tok := handler.client.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		mqttLogger.WithError(tok.Error()).Fatal("Could not connect to mqtt server.")
	}

	return &handler
}

func (h *MqttHandler) CurrentStatus() string {
	return h.status
}

func (h *MqttHandler) SendDoorBuzzer(door Door) bool {
	if h.status != "open" && h.status != "open+" && h.status != "member" {
		mqttLogger.WithField("status", h.status).Error("door buzzer is not allowed for the current status.")
		return false
	}

	var topic string
	switch door {
	case DoorOuter:
		topic = h.conf.DoorDownstairsBuzzerTopic
		break
	case DoorInnerGlass:
		topic = h.conf.GlassDoorBuzzerTopic
		break
	case DoorInnerMetal:
		topic = h.conf.MainDoorBuzzerTopic
		break
	}
	token := h.client.Publish(topic, 0, false, fmt.Sprintf("%d", BUZZER_DURATION))
	ok := token.WaitTimeout(time.Duration(time.Second * 10))
	if !ok {
		mqttLogger.WithError(token.Error()).WithField("topic", topic).Info("Error sending door buzzer.")
		return false
	}

	return true
}

func (h *MqttHandler) onConnect(client mqtt.Client) {
	mqttLogger.Info("connected")

	err := subscribe(client, h.status,
		func(client mqtt.Client, message mqtt.Message) {
			h.status = string(message.Payload())
			mqttLogger.WithField("status", h.status).Info("got new status")
		})
	if err != nil {
		mqttLogger.WithError(err).Fatal("Could not subscribe.")
	}
}

func (h *MqttHandler) onConnectionLost(client mqtt.Client, err error) {
	mqttLogger.WithError(err).Error("Connection lost.")
	// clearing the status
	h.status = ""
}

func subscribe(client mqtt.Client, topic string, cb mqtt.MessageHandler) error {
	qos := 0
	tok := client.Subscribe(topic, byte(qos), cb)
	tok.WaitTimeout(5 * time.Second)
	return tok.Error()
}

func defaultCertPool(certFile string) *x509.CertPool {
	if certFile == "" {
		mqttLogger.Debug("No certFile given, using system pool")
		pool, err := x509.SystemCertPool()
		if err != nil {
			mqttLogger.WithError(err).Fatal("Could not create system cert pool.")
		}
		return pool
	}

	fileData, err := ioutil.ReadFile(certFile)
	if err != nil {
		mqttLogger.WithError(err).Fatal("Could not read given cert file.")
	}

	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(fileData) {
		mqttLogger.Fatal("unable to add given certificate to CertPool")
	}

	return certs
}
