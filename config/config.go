package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/goburrow/serial"
)

type MqttConfig struct {
	Enabled  bool
	Url      string
	Topic    string
	Qos      byte
	Insecure bool
}

type ModbusSerialConfig struct {
	Enabled bool
	Port    serial.Config
}

type ModbusTcpConfig struct {
	Host string
	Port int
}

type ModbusConfig struct {
	Serial ModbusSerialConfig
	Tcp    ModbusTcpConfig
}

type Config struct {
	Debug  bool
	Mqtt   MqttConfig
	Modbus ModbusConfig
}

var Conf Config

func Load() {
	// TODO: use https://github.com/spf13/viper
	if _, err := os.Stat("/etc/guacamole.conf"); os.IsNotExist(err) {
		if _, err := toml.DecodeFile("./conf.d/guacamole.conf", &Conf); err != nil {
			log.Fatal("[Config] TOML parse error: ", err.Error())
		}
	} else {
		if _, err := toml.DecodeFile("/etc/guacamole.conf", &Conf); err != nil {
			log.Fatal("[Config] TOML parse error: ", err.Error())
		}
	}
}

func HasMqtt() (result bool) {
	result = Conf.Mqtt.Enabled
	return
}
