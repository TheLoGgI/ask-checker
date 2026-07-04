package main

import (
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port         string `env:"APP_PORT" envDefault:"3000"`
	DATABASE_URL string `env:"DATABASE_URL"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
	Environment  string `env:"APP_ENV" envDefault:"local"`
	Database     string `env:"DATABASE_TYPE"`
}

var Cfg Config

func init() {
	if err := env.Parse(&Cfg); err != nil {
		log.Fatalf("Missing required config: %v", err)
	}

	if Cfg.Database == "" {
		isVercel := strings.EqualFold(os.Getenv("VERCEL"), "1") || strings.EqualFold(os.Getenv("VERCEL_ENV"), "production")
		if strings.EqualFold(Cfg.Environment, "production") || isVercel {
			Cfg.Database = "postgres"
		} else {
			Cfg.Database = "sqlite"
		}
	}

	if strings.EqualFold(Cfg.Database, "postgres") && Cfg.DATABASE_URL == "" {
		log.Fatal("DATABASE_URL is required when DATABASE_TYPE=postgres")
	}
}
