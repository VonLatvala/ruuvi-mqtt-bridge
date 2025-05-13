# Ruuvi MQTT Bridge

![Build Status](https://img.shields.io/github/actions/workflow/status/VonLatvala/ruuvi-mqtt-bridge/release-please.yml)
![Release Version](https://img.shields.io/github/v/release/VonLatvala/ruuvi-mqtt-bridge)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/VonLatvala/ruuvi-mqtt-bridge)
![License](https://img.shields.io/github/license/VonLatvala/ruuvi-mqtt-bridge)

## Overview

**Ruuvi MQTT Bridge** is a Go-based application that bridges data from Scrin's RuuviCollector to an MQTT broker. This enables seamless integration of RuuviTag sensor data into IoT ecosystems, facilitating real-time monitoring and automation.

## Features

* **Data Bridging**: Transfers sensor data from RuuviCollector to MQTT.
* **Real-time Updates**: Publishes sensor data to MQTT topics in real-time.
* **Configurable**: Easily configurable to connect to different MQTT brokers and RuuviCollector instances.

## Prerequisites

Before running the Ruuvi MQTT Bridge, ensure you have the following:

* Go 1.24 or higher
* Docker (for building within a containerized environment)
* Make (for managing build tasks)

## Installation

### Option 1: Build from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/VonLatvala/ruuvi-mqtt-bridge.git
   cd ruuvi-mqtt-bridge
   ```

2. Build the application:

   ```bash
   make build-container && make build-binary
   ```

3. Install the application:

   ```bash
   sudo make install
   ```

### Option 2: Using Docker

1. Build the Docker image:

   ```bash
   docker build -f .devcontainer/Dockerfile -t ruuvi-mqtt-bridge .
   ```

2. Run the Docker container:

   ```bash
   docker run --rm -it \
     -v "$(pwd)":/workspace \
     -w /workspace \
     ruuvi-mqtt-bridge \
     ./bin/ruuvi-mqtt-bridge-linux-amd64
   ```

## Configuration

The application supports configuration via **environment variables** or **command-line flags**. Command-line flags always override environment variables, which in turn override the default values.

### ⚙️ Configuration Options

| Option               | Env Var              | CLI Flag              | Default Value            | Description                                                    |
| -------------------- | -------------------- | --------------------- | ------------------------ | -------------------------------------------------------------- |
| Discovery Prefix     | `DISCOVERY_PREFIX`   | `-discovery-prefix`   | `homeassistant`          | Home Assistant MQTT discovery prefix.                          |
| InfluxDB URL         | `INFLUX_URL`         | `-influx-url`         | `http://localhost:8086`  | Base URL of your InfluxDB instance.                            |
| InfluxDB Name        | `INFLUX_DB`          | `-influx-db`          | `ruuvi`                  | Name of the InfluxDB database.                                 |
| InfluxDB Measurement | `INFLUX_MEASUREMENT` | `-influx-measurement` | `ruuvi_measurements`     | InfluxDB measurement name.                                     |
| MQTT Host            | `MQTT_HOST`          | `-mqtt-host`          | `localhost`              | Hostname or IP of the MQTT broker.                             |
| MQTT Port            | `MQTT_PORT`          | `-mqtt-port`          | `1883`                   | MQTT broker port.                                              |
| MQTT Username        | `MQTT_USER`          | `-mqtt-user`          | *(none)*                 | Username for MQTT broker.                                      |
| MQTT Password        | `MQTT_PASS`          | `-mqtt-pass`          | *(none)*                 | Password for MQTT broker.                                      |
| MQTT Topic Prefix    | `MQTT_TOPIC_PREFIX`  | `-mqtt-topic-prefix`  | `ruuvi`                  | Base topic under which to publish data.                        |
| Properties File      | `PROPERTIES_FILE`    | `-properties-file`    | `ruuvi-names.properties` | Path to `ruuvi-names.properties` used for sensor naming.       |
| Scrape Interval      | `SCRAPE_INTERVAL`    | `-scrape-interval`    | `5s`                     | How often to poll RuuviCollector for data (e.g., `10s`, `1m`). |

### 🧪 Example: Using Environment Variables

Create a `.env` file or export variables in your shell:

```bash
export MQTT_HOST=broker.local
export MQTT_USER=myuser
export MQTT_PASS=mypassword
export INFLUX_URL=http://influxdb.local:8086
export SCRAPE_INTERVAL=10s
```

Then run the app:

```bash
./ruuvi-mqtt-bridge-linux-amd64
```

### 🚀 Example: Using Command-Line Flags

```bash
./ruuvi-mqtt-bridge-linux-amd64 \
  -mqtt-host broker.local \
  -mqtt-user myuser \
  -mqtt-pass mypassword \
  -influx-url http://influxdb.local:8086 \
  -scrape-interval 10s
```

### 🔄 Precedence

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Default values** (lowest priority)

## Releases

Releases are automatically created upon merging pull requests into the `main` branch. The release pipeline builds the application and uploads the binaries to the GitHub release associated with the tag.

To download a release:

1. Visit the [Releases](https://github.com/VonLatvala/ruuvi-mqtt-bridge/releases) page.
2. Download the appropriate binary for your platform.

## Contributing

We welcome contributions! To contribute:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/YourFeature`.
3. Make your changes and commit them: `git commit -m 'Add some feature'`.
4. Push to the branch: `git push origin feature/YourFeature`.
5. Open a Pull Request.

Please ensure your code adheres to the existing coding style and includes appropriate tests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
