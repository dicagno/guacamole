<img src=".github/assets/guacamole-logo.jpg" width="256" />

# Guacamole - seasoning for automation

## Description

This tool connects any PLCs supporting MODBUS TCP or RTU to an MQTT broker, effectively enabling realtime messaging in already existing devices without impacting them.
It comes from a project from 2019, but still useful in several contexts.

## Features

* Emulates a MODBUS slave
* Low footprint
* Can be run either on a gateway or on the PLC itself
* Cross-platform, cross-architecture
* Supports both TCP and Serial MODBUS masters
* Forwards register write requests (either single, `FC 6` or multiple, `FC 16`) to the configured PubSub broker (and topic)

## Building artifacts

```bash
make
```

## Installation

```bash
make install
```

Compile for different architectures, i.e. MIPS Little Endian

```bash
GOOS=linux GOARCH=mipsle go build -o ./release/guacamole-mipsel main.go
```

## Installation on remote device

This method makes use of SSH, key-based auth

```bash
make deploy USER=myuser IP=192.168.1.1
```

## Usage

```bash
service modbus-mqtt enable
service modbus-mqtt start
```