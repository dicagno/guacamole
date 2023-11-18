package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dicagno/guacamole/config"
	"github.com/dicagno/guacamole/messaging"
	"github.com/dicagno/guacamole/packet"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/tbrandon/mbserver"
)

var errorCount uint16
var mqttClient *mqtt.Client

// BytesToUint16 converts a big endian array of bytes to an array of unit16s
func BytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])
	}
	return values
}

func BytesToInt32(bytes []byte) []int32 {
	values := make([]int32, len(bytes)/2)

	for i := range values {
		values[i] = int32(binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2]))
	}
	return values
}

func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

func BitAtPosition(value uint8, pos uint) uint8 {
	return (value >> pos) & 0x01
}

func initModbus() {
	ModbusServer := mbserver.NewServer()
	ModbusServer.RegisterFunctionHandler(6, WriteHoldingRegisterCallback)
	ModbusServer.RegisterFunctionHandler(16, WriteHoldingRegistersCallback)

	var ModbusServerError error
	if config.Conf.Modbus.Serial.Enabled {
		//ModbusServerError = ModbusServer.ListenRTU(&config.Conf.Modbus.Serial.Port)
		log.Fatal("Modbus RTU not yet implemented.")
	} else {
		ModbusServerError = ModbusServer.ListenTCP(config.Conf.Modbus.Tcp.Host + ":" + string(config.Conf.Modbus.Tcp.Host))
	}
	if ModbusServerError != nil {
		log.Fatal(fmt.Printf("[modbus_server_init] mberr: %v\n", ModbusServerError))
	}

	defer ModbusServer.Close()
}

func registerAddressAndNumber(frame mbserver.Framer) (register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	endRegister = register + numRegs
	return register, numRegs, endRegister
}

func registerAddressAndValue(frame mbserver.Framer) (int, uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))
	value := binary.BigEndian.Uint16(data[2:4])
	return register, value
}

func WriteHoldingRegistersCallback(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	fmt.Print("WriteHoldingRegistersCallback", frame)
	register, numRegs, _ := registerAddressAndNumber(frame)
	valueBytes := frame.GetData()[5:]
	var exception *mbserver.Exception
	var data []byte

	if len(valueBytes)/2 != numRegs {
		exception = &mbserver.IllegalDataAddress
	}

	values := BytesToInt32(valueBytes)
	p := new(packet.Packet)
	p.FunctionCode = packet.FunctionCode(frame.GetFunction())
	p.RegisterAddress = int32(register)
	p.RegistersCount = int32(numRegs)
	p.RegisterValues = make([]int32, numRegs)
	valuesUpdated := copy(p.RegisterValues /*dst*/, values /*src*/)

	if valuesUpdated == numRegs {
		exception = &mbserver.Success
		data = frame.GetData()[0:4]
	} else {
		exception = &mbserver.IllegalDataAddress
	}

	println("packet = ", p.String())

	DeliverPacket(p)

	return data, exception
}

func WriteHoldingRegisterCallback(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	register, value := registerAddressAndValue(frame)

	p := new(packet.Packet)
	p.FunctionCode = packet.FunctionCode(frame.GetFunction())
	p.RegisterValue = int32(value)
	p.RegistersCount = int32(1)
	p.RegisterAddress = int32(register)

	println("packet = ", p.String())

	DeliverPacket(p)

	return frame.GetData()[0:4], &mbserver.Success
}

func DeliverPacket(pk *packet.Packet) {
	var marshalErr error
	var data = new([]byte)
	*data, marshalErr = proto.Marshal(pk)
	if marshalErr != nil {
		log.Fatal("marshaling error: ", marshalErr)
	}
	//fmt.Println("proto_data=", data)
	messaging.PublishViaMqtt(data)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	config.Load()

	ModbusServer := mbserver.NewServer()
	ModbusServer.RegisterFunctionHandler(6, WriteHoldingRegisterCallback)
	ModbusServer.RegisterFunctionHandler(16, WriteHoldingRegistersCallback)

	var ModbusServerError error
	var ModbusListen = new(string)
	if config.Conf.Modbus.Serial.Enabled {
		ModbusServerError = ModbusServer.ListenRTU(&config.Conf.Modbus.Serial.Port)
	} else {
		*ModbusListen = fmt.Sprintf("%s:%d", config.Conf.Modbus.Tcp.Host, config.Conf.Modbus.Tcp.Port)
		ModbusServerError = ModbusServer.ListenTCP(*ModbusListen)
	}
	if ModbusServerError != nil {
		log.Fatal(fmt.Printf("[MODBUS_INIT] mberr: %v\n", ModbusServerError))
	}

	//defer ModbusServer.Close()

	messaging.ConnectToMqtt()
	//defer messaging.MqttClient.Disconnect(1)

	for {
		time.Sleep(time.Millisecond * (1000 ^ 4))
	}
}
