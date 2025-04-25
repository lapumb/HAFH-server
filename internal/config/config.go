package config

import (
	"os"

	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug bool        `yaml:"debug" default:"false"`
	HTTP  HTTPConfig  `yaml:"http"`
	Ngrok NgrokConfig `yaml:"ngrok"`
	MQTT  MQTTConfig  `yaml:"mqtt"`
	DB    DBConfig    `yaml:"database"`
}

type HTTPConfig struct {
	Port                 int    `yaml:"port" default:"8080"`
	APIKey               string `yaml:"api_key" default:""`
	MaxRequestsPerSecond int    `yaml:"max_requests_per_second" default:"5"`
}

type NgrokConfig struct {
	Enabled   bool   `yaml:"enabled" default:"false"`
	AuthToken string `yaml:"auth_token" default:""`
	Domain    string `yaml:"domain" default:""`
	Region    string `yaml:"region" default:"us"`
}

type MQTTConfig struct {
	Address  string `yaml:"address" default:"0.0.0.0"`
	Port     int    `yaml:"port" default:"8883"`
	CertPath string `yaml:"cert_path" default:"certs/server.crt"`
	KeyPath  string `yaml:"key_path" default:"certs/server.key"`
	CaPath   string `yaml:"ca_path" default:"certs/ca.crt"`
}

type DBConfig struct {
	Path string `yaml:"path" default:":memory:"`
}

// String returns the string representation of the Config struct.
func (c *Config) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}

	return string(data)
}

// Load reads the configuration from the specified YAML file and populates the Config struct.
func Load(path string) (*Config, error) {
	config := &Config{}
	defaults.SetDefaults(config)

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
