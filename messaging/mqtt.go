package messaging

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dicagno/guacamole/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MqttClient mqtt.Client

func ConnectToMqtt() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	clientId := fmt.Sprintf("%s_modbus-%d", hostname, rand.Int())
	opts := CreateMqttClientOptions(clientId, config.Conf.Mqtt.Url)
	MqttClient = mqtt.NewClient(opts)
	token := MqttClient.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
}

func OnMqttConnectionLost(_ mqtt.Client, e error) {
	if config.Conf.Debug {
		log.Println("[MQTT] connection lost: ", e)
	}
}

func CreateMqttClientOptions(clientId string, brokerSocketString string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.SetClientID(clientId)
	opts.AddBroker(brokerSocketString)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(10)
	if strings.HasPrefix(config.Conf.Mqtt.Url, "ssl://") && config.Conf.Mqtt.Insecure {
		if config.Conf.Debug {
			log.Println("[MQTT] ssl INSECURE (skipping cert validation)")
		}
		t := tls.Config{
			InsecureSkipVerify: true,
		}
		opts.SetTLSConfig(&t)
	}

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		if config.Conf.Debug {
			log.Println("[MQTT] client connected")
		}
	})
	opts.SetConnectionLostHandler(OnMqttConnectionLost)
	return opts
}

func PublishViaMqtt(pkt *[]byte) {
	MqttClient.Publish(config.Conf.Mqtt.Topic, config.Conf.Mqtt.Qos, false, *pkt)
}
