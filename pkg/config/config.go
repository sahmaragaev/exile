package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	MongoURI      string `json:"mongo_uri"`
	OpenAIKey     string `json:"openai_api_key"`
	AssistantId   string `json:"assistant_id"`
	TelegramToken string `json:"telegram_token"`
}

var AppConfig Config

func LoadConfig(configFile string) error {
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		return err
	}

	return nil
}
