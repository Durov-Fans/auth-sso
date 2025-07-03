package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env          string        `yaml:"env" env-default:"local"`
	Database_url string        `yaml:"database_url" env-required:"true"`
	GRPC         GRPCConfig    `yaml:"grpc" env-required:"true"`
	TokenTTL     time.Duration `yaml:"token_ttl" env-default:"5h"`
}

type GRPCConfig struct {
	Port    int    `yaml:"port" env-default:"8080"`
	Timeout string `yaml:"timeout" env-default:"5h"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config file path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found")
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic(err)
	}

	return &config
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	fmt.Println(res)
	return res

}
