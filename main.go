package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v9"
)

var pgCfg postgresConfig
var rCfg redisConfig
var tgCfg telegramConfig

var dbInfo string

var rdb *redis.Client

func main() {
	if err := env.Parse(&pgCfg); err != nil {
		log.Fatalln("[Postgres env config] FATAL:", err)
	}
	dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pgCfg.Host, pgCfg.Port, pgCfg.User, pgCfg.Pass, pgCfg.DBName, pgCfg.SSLMode)

	if err := env.Parse(&rCfg); err != nil {
		log.Fatalln("[Redis env config] FATAL:", err)
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     rCfg.Addr,
		Password: rCfg.Password,
		DB:       rCfg.DB,
	})

	if err := env.Parse(&tgCfg); err != nil {
		log.Fatalln("[Telegram env config] FATAL:", err)
	}

	if err := CreateTablePG(); err != nil {
		log.Fatalln("[Create table] FATAL:", err)
	}

	startBot(tgCfg.TelegramToken)
}
