package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	kitlog "github.com/go-kit/log"
	kitloglevel "github.com/go-kit/log/level"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Configurable variables from env or flags
var (
	influxURL         = getEnvOrFlag("INFLUX_URL", "influx-url", "http://localhost:8086", "InfluxDB base URL")
	influxDB          = getEnvOrFlag("INFLUX_DB", "influx-db", "ruuvi", "InfluxDB database name")
	influxMeasurement = getEnvOrFlag("INFLUX_MEASUREMENT", "influx-measurement", "ruuvi_measurements", "InfluxDB measurement")
	mqttHost          = getEnvOrFlag("MQTT_HOST", "mqtt-host", "localhost", "MQTT broker host")
	mqttPort          = getEnvOrFlagInt("MQTT_PORT", "mqtt-port", 1883, "MQTT broker port")
	mqttUser          = getEnvOrFlag("MQTT_USER", "mqtt-user", "", "MQTT username")
	mqttPass          = getEnvOrFlag("MQTT_PASS", "mqtt-pass", "", "MQTT password")
	mqttTopicPrefix   = getEnvOrFlag("MQTT_TOPIC_PREFIX", "mqtt-topic-prefix", "ruuvi", "Base MQTT topic")
	discoveryPrefix   = getEnvOrFlag("DISCOVERY_PREFIX", "discovery-prefix", "homeassistant", "Home Assistant discovery prefix")
	scrapeInterval    = getEnvOrFlagDuration("SCRAPE_INTERVAL", "scrape-interval", 5*time.Second, "Scrape interval")
	propertiesFile    = getEnvOrFlag("PROPERTIES_FILE", "properties-file", "ruuvi-names.properties", "Path to ruuvi-names.properties")
)

var measurementFields = []string{
	"temperature", "humidity", "pressure", "battery", "rssi", "dewPoint",
	"accelX", "accelY", "accelZ", "movementCounter",
	"accelAngleX", "accelAngleY", "accelAngleZ", "accelTotal",
	"accelStatus", "accelMotion", "accelCounter", "accelSequence", "accelTimestamp",
	"accelRawX", "accelRawY", "accelRawZ", "accelRawTotal",
	"accelRawStatus", "accelRawMotion", "accelRawCounter",
	"accelRawSequence", "accelRawTimestamp",
}

var fieldUnits = map[string]string{
	"temperature":       "°C",
	"humidity":          "%",
	"pressure":          "hPa",
	"battery":           "V",
	"rssi":              "dBm",
	"dewPoint":          "°C",
	"accelX":            "g",
	"accelY":            "g",
	"accelZ":            "g",
	"accelTotal":        "g",
	"accelAngleX":       "°",
	"accelAngleY":       "°",
	"accelAngleZ":       "°",
	"accelRawX":         "g",
	"accelRawY":         "g",
	"accelRawZ":         "g",
	"accelRawTotal":     "g",
	"movementCounter":   "count",
	"accelCounter":      "count",
	"accelSequence":     "",
	"accelTimestamp":    "s",
	"accelStatus":       "",
	"accelMotion":       "",
	"accelRawCounter":   "count",
	"accelRawSequence":  "",
	"accelRawTimestamp": "s",
	"accelRawStatus":    "",
	"accelRawMotion":    "",
}

var fieldDeviceClasses = map[string]string{
	"temperature":     "temperature",
	"humidity":        "humidity",
	"pressure":        "pressure",
	"battery":         "battery",
	"rssi":            "signal_strength",
	"dewPoint":        "temperature",
	"movementCounter": "motion",
	"accelCounter":    "counter",
	"accelRawCounter": "counter",
}

var fieldIcons = map[string]string{
	"accelX":        "mdi:axis-x-arrow",
	"accelY":        "mdi:axis-y-arrow",
	"accelZ":        "mdi:axis-z-arrow",
	"accelAngleX":   "mdi:rotate-left",
	"accelAngleY":   "mdi:rotate-left",
	"accelAngleZ":   "mdi:rotate-left",
	"accelTotal":    "mdi:vibrate",
	"accelMotion":   "mdi:run-fast",
	"accelStatus":   "mdi:information",
	"accelSequence": "mdi:format-list-numbered",
	"accelTimestamp": "mdi:clock-outline",
	"accelRawX":        "mdi:axis-x-arrow",
	"accelRawY":        "mdi:axis-y-arrow",
	"accelRawZ":        "mdi:axis-z-arrow",
	"accelRawTotal":    "mdi:vibrate",
	"accelRawMotion":   "mdi:run-fast",
	"accelRawStatus":   "mdi:information",
	"accelRawSequence": "mdi:format-list-numbered",
	"accelRawTimestamp": "mdi:clock-outline",
}

var logger kitlog.Logger



func main() {
	logger = kitlog.NewLogfmtLogger(os.Stdout)
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)

	flag.Parse()

	kitloglevel.Info(logger).Log("msg", "Loading MAC → name mappings...")
	names, err := loadProperties(*propertiesFile)
	if err != nil {
		kitloglevel.Error(logger).Log("msg", "Failed to read properties file", "err", err)
		os.Exit(1)
	}
	kitloglevel.Info(logger).Log("msg", "Loaded MAC name mappings", "num", len(names))

	kitloglevel.Info(logger).Log("msg", "Connecting to MQTT broker...")
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", *mqttHost, *mqttPort))
	opts.SetClientID("ruuvi-mqtt-bridge")
	if *mqttUser != "" {
		opts.SetUsername(*mqttUser)
		opts.SetPassword(*mqttPass)
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		kitloglevel.Error(logger).Log("msg", "Failed to connect to MQTT", "err", err)
		os.Exit(1)
	}
	defer client.Disconnect(250)
	kitloglevel.Info(logger).Log("msg", "Connected to MQTT.")

	kitloglevel.Info(logger).Log("msg", "Sending Home Assistant discovery configs...")
	sendDiscoveryConfigs(client, names)

	for {
		data, err := queryLatestInfluxWithBackoff()
		if err != nil {
			kitloglevel.Error(logger).Log("msg", "Final failure querying InfluxDB", "err", err)
			time.Sleep(*scrapeInterval)
			continue
		}

		kitloglevel.Info(logger).Log("msg", "Retrieved MAC entries from InfluxDB", "count", len(data))
		for mac, fields := range data {
			name := names[mac]
			if name == "" {
				name = mac
			}
			topic := fmt.Sprintf("%s/%s", *mqttTopicPrefix, name)
			payload, _ := json.Marshal(fields)
			client.Publish(topic, 0, true, payload)
		}
		kitloglevel.Info(logger).Log("msg", "Published MQTT messages", "num", len(data))
		time.Sleep(*scrapeInterval)
	}
}

func queryLatestInfluxWithBackoff() (map[string]map[string]interface{}, error) {
	const maxRetries = 5
	backoff := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		data, err := executeInfluxQuery()
		if err == nil {
			return data, nil
		}
		kitloglevel.Error(logger).Log("msg", "InfluxDB query failed. Retrying...", "attempt", i+1, "maxRetries", maxRetries, "err", err, "backoff", backoff)
		time.Sleep(backoff)
		backoff *= 2
	}
	return nil, fmt.Errorf("InfluxDB query failed after %d attempts", maxRetries)
}

func executeInfluxQuery() (map[string]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT
			LAST(temperature) AS temperature,
			LAST(humidity) AS humidity,
			LAST(pressure) AS pressure,
			LAST(battery) AS battery,
			LAST(rssi) AS rssi,
			LAST(dewPoint) AS dewPoint,
			LAST(accelX) AS accelX,
			LAST(accelY) AS accelY,
			LAST(accelZ) AS accelZ,
			LAST(movementCounter) AS movementCounter,
			LAST(accelAngleX) AS accelAngleX,
			LAST(accelAngleY) AS accelAngleY,
			LAST(accelAngleZ) AS accelAngleZ,
			LAST(accelTotal) AS accelTotal,
			LAST(accelStatus) AS accelStatus,
			LAST(accelMotion) AS accelMotion,
			LAST(accelCounter) AS accelCounter,
			LAST(accelSequence) AS accelSequence,
			LAST(accelTimestamp) AS accelTimestamp,
			LAST(accelRawX) AS accelRawX,
			LAST(accelRawY) AS accelRawY,
			LAST(accelRawZ) AS accelRawZ,
			LAST(accelRawTotal) AS accelRawTotal,
			LAST(accelRawStatus) AS accelRawStatus,
			LAST(accelRawMotion) AS accelRawMotion,
			LAST(accelRawCounter) AS accelRawCounter,
			LAST(accelRawSequence) AS accelRawSequence,
			LAST(accelRawTimestamp) AS accelRawTimestamp
		FROM "%s"
		WHERE time > now() - 10m
		GROUP BY "mac"
	`, *influxMeasurement)

	queryURL := fmt.Sprintf("%s/query?db=%s&q=%s", *influxURL, *influxDB, url.QueryEscape(query))
	kitloglevel.Info(logger).Log("msg", "InfluxDB Query URL", queryURL)

	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var result struct {
		Results []struct {
			Series []struct {
				Tags    map[string]string `json:"tags"`
				Columns []string          `json:"columns"`
				Values  [][]interface{}   `json:"values"`
			} `json:"series"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}

	data := map[string]map[string]interface{}{}
	for _, res := range result.Results {
		for _, series := range res.Series {
			mac := series.Tags["mac"]
			if len(series.Values) == 0 {
				continue
			}
			row := series.Values[0]
			fields := map[string]interface{}{}
			for i, col := range series.Columns {
				if col == "time" {
					continue
				}
				fields[col] = row[i]
			}
			data[mac] = fields
		}
	}
	return data, nil
}

func sendDiscoveryConfigs(client mqtt.Client, names map[string]string) {
	for _, field := range measurementFields {
		for mac, name := range names {
			if name == "" {
				name = mac
			}
			baseID := strings.ReplaceAll(name, " ", "_")
			id := fmt.Sprintf("ruuvi_%s_%s", baseID, field)
			configTopic := fmt.Sprintf("%s/sensor/%s/config", *discoveryPrefix, id)
			stateTopic := fmt.Sprintf("%s/%s", *mqttTopicPrefix, name)
			config := map[string]interface{}{
				"name":            fmt.Sprintf("%s %s", name, strings.Title(field)),
				"state_topic":     stateTopic,
				"value_template":  fmt.Sprintf("{{ value_json.%s }}", field),
				"unique_id":       id,
				"device": map[string]interface{}{
					"identifiers":  []string{"ruuvi_" + baseID},
					"name":         name,
					"manufacturer": "Ruuvi",
					"model":        "RuuviTag",
				},
			}

			if unit, ok := fieldUnits[field]; ok && unit != "" {
				config["unit_of_measurement"] = unit
			}
			if devClass, ok := fieldDeviceClasses[field]; ok {
				config["device_class"] = devClass
			}
			if icon, ok := fieldIcons[field]; ok {
				config["icon"] = icon
			}
			payload, _ := json.Marshal(config)
			client.Publish(configTopic, 0, true, payload)
		}
	}
}

func loadProperties(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	props := make(map[string]string)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		props[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return props, nil
}

// Helpers for environment + CLI config
func getEnvOrFlag(env, name, def, desc string) *string {
	val := os.Getenv(env)
	return flag.String(name, ifNotEmpty(val, def), desc)
}

func getEnvOrFlagInt(env, name string, def int, desc string) *int {
	val := os.Getenv(env)
	if val != "" {
		var parsed int
		_, err := fmt.Sscanf(val, "%d", &parsed)
		if err == nil {
			return flag.Int(name, parsed, desc)
		}
		kitloglevel.Warn(logger).Log("msg", "Warning: could not parse attemptedValue as int, using default", "attemptedValue", val, "default", def)
	}
	return flag.Int(name, def, desc)
}

func getEnvOrFlagDuration(env, name string, def time.Duration, desc string) *time.Duration {
	val := os.Getenv(env)
	if val != "" {
		parsed, err := time.ParseDuration(val)
		if err == nil {
			return flag.Duration(name, parsed, desc)
		}
		kitloglevel.Warn(logger).Log("msg", "Warning: could not parse attemptedValue as duration, using default", "attemptedValue", val, "default", def)
	}
	return flag.Duration(name, def, desc)
}

func ifNotEmpty(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}
