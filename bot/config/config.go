package config

import "github.com/caarlos0/env/v6"

type Cfg struct {
	Token           string `env:"BOT_TOKEN,required"`
	ElectricityArea string `env:"EL_AREA,required"`
	ChatsToNotify   []int  `env:"CHATS_TO_NOTIFY"`
}

func Get() (*Cfg, error) {
	cfg := Cfg{}

	err := env.Parse(&cfg)

	return &cfg, err
}
