package main

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func messageHandler(bot tgbotapi.BotAPI, upd tgbotapi.Update) {
	redis, err := GetStateRedis(upd.Message.From.UserName, upd.Message.From.ID)
	if err != nil {
		log.Println("[Get state redis] ERROR:", err)
		return
	}

	switch upd.Message.Command() {
	case "start":
		ProcessSrartOperation(bot, upd)
		return
	case "help":
		ProcessHelpOperation(bot, upd)
		return
	case "cancel":
		ProcessCancelOpearation(bot, upd, redis)
		return
	}

	if redis.State == Def {
		switch upd.Message.Command() {
		case "add":
			go InitAddingSticker(bot, upd, redis)
		case "delete":
			go InitDeletingSticker(bot, upd, redis)
		case "show":
			go InitShowDescription(bot, upd, redis)
		default:
			go ProcessUnrecognizedCommand(bot, upd)
		}
	} else {
		go ProcessFSM(bot, upd, redis)
	}
}

func inlineQueryHandler(bot tgbotapi.BotAPI, upd tgbotapi.Update) {
	userID := upd.InlineQuery.From.ID
	query := upd.InlineQuery.Query
	queryOffset, _ := strconv.Atoi(upd.InlineQuery.Offset)

	rows, err := GetStickerPG(userID, query, queryOffset)
	if err != nil {
		log.Println("[Get Stickers] ERROR:", err)
	}

	var results []any
	for _, row := range rows {
		sticker := tgbotapi.NewInlineQueryResultCachedSticker(row.stickerID[:64], row.stickerID, upd.InlineQuery.Query)
		results = append(results, sticker)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: upd.InlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     10,
		Results:       results,
	}

	if len(results) == 50 {
		inlineConf.NextOffset = strconv.Itoa(queryOffset + 50)
	}

	if _, err := bot.Request(inlineConf); err != nil {
		log.Println("[Inline request] ERROR:", err)
	}
}

func startBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalln("[Telegram сonnection] FATAL:", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		switch {
		case update.Message != nil:
			go messageHandler(*bot, update)
		case update.InlineQuery != nil:
			go inlineQueryHandler(*bot, update)
		}
	}
}
