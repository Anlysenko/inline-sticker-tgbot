package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/caarlos0/env/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

type config struct {
	pgdb struct {
		DSN          string `env:"PG_DSN,required"`
		MaxOpenConns int    `env:"PG_MaxOpenConns,required"`
		MaxIdleConns int    `env:"PG_MaxIdleConns,required"`
		MaxIdleTime  string `env:"PG_MaxIdleTime,required"`
	}

	telegram struct {
		Token string `env:"TG_TOKEN,required"`
	}
}

type application struct {
	sticker    StickerModel
	user       UserModel
	telegram   *TelegramModel
	fsm        *StateMachine
	memSticker *MemSticker
}

func main() {
	var cfg config

	if err := env.Parse(&cfg.pgdb); err != nil {
		log.Fatalln("[Postgres env config] FATAL:", err)
	}

	db, err := openDB(cfg)
	if err != nil {
		log.Fatalln("[Open DB] FATAL:", err)
	}
	defer db.Close()

	if err := env.Parse(&cfg.telegram); err != nil {
		log.Fatalln("[Telegram env config] FATAL:", err)
	}

	tg, err := initTelegramBot(cfg.telegram.Token)
	if err != nil {
		log.Fatalln("[Telegram connection] FATAL:", err)
	}

	app := &application{
		sticker:    StickerModel{DB: db},
		user:       UserModel{DB: db},
		telegram:   tg,
		fsm:        newStickerFSM(),
		memSticker: newMemSticker(),
	}

	go app.memSticker.backgroundCleanup()

	app.startBot()
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.pgdb.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.pgdb.MaxOpenConns)
	db.SetMaxIdleConns(cfg.pgdb.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.pgdb.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initTelegramBot(token string) (*TelegramModel, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	tg := &TelegramModel{bot: bot}

	return tg, nil
}
