package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	config Config
)

type Config struct {
	Twitter  TwitterConfig  `yaml:"twitter"`
	Telegram TelegramConfig `yaml:"telegram"`
}

type TwitterConfig struct {
	CustomerKey       string `yaml:"customer_key"`
	CustomerSecret    string `yaml:"customer_secret"`
	AccessToken       string `yaml:"access_token"`
	AccessTokenSecret string `yaml:"access_token_secret"`
}

type TelegramConfig struct {
	TGBotToken string `yaml:"tg_bot_token"`
	ChannelID  string `yaml:"channel_id"`
}

func LoadConfig() Config {
	yamlFile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		panic(err)
	}
	return config
}

func StartConfig() {
	LoadConfig()
}

func GetConfig() Config {
	return config
}
