package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var _config *Config

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Rollbar  RollbarConfig  `json:"rollbar"`
	Redis    RedisConfig    `json:"redis"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
	Gmail    GmailConfig    `json:"gmail"`
}

type ServerConfig struct {
	Mode            string `json:"mode"`
	Port            int    `json:"port"`
	Debug           bool   `json:"debug,omitempty"`
	HostUrl         string `json:"host_url"`
	ApiUrl          string `json:"api_url"`
	JWTSecret       string `json:"jwt_secret"`
	JWTHeader       string `json:"jwt_header"`
	JWTBearerPrefix string `json:"jwt_bearer_prefix"`
}

type RollbarConfig struct {
	Environment string `json:"environment"`
	Token       string `json:"token"`
	ServerRoot  string `json:"server_root"`
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Host     string `json:"host"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

type RabbitMQConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type GmailConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RefreshToken string   `json:"refresh_token"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

func NewConfig(filePath string) *Config {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("can't open config file: %s", err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("can't read config file: %s", err)
	}

	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		currentPath, _ := os.Getwd()
		log.Fatalf("config file unmarshalling error: %s\ncurrent wd: %s\nconfiguration path:%s", err, currentPath, filePath)
	}

	_config = &config
	return &config
}

func GetConfig() *Config {
	return _config
}
