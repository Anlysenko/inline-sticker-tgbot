package main

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramModel struct {
	bot *tgbotapi.BotAPI
	upd *tgbotapi.Update
}

func (app *application) startBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := app.telegram.bot.GetUpdatesChan(u)

	for update := range updates {
		app.telegram.upd = &update
		switch {
		case update.Message != nil:
			app.messageHandler(update.Message.Text)
		case update.InlineQuery != nil:
			app.inlineQueryHandler(update)
		}
	}
}

func (app *application) messageHandler(message string) {
	switch message {
	case "/start":
		app.startOperation()
		return
	case "/cancel":
		app.cancelOperation()
		return
	case "/help":
		app.telegram.helpOperationResponse()
		return
	}

	actionContext := &ActionContext{
		tgModel:    app.telegram,
		sticker:    app.sticker,
		user:       app.user,
		memSticker: app.memSticker,
	}

	u := &User{
		ID:    app.telegram.upd.Message.From.ID,
		Name:  app.telegram.upd.Message.From.UserName,
		State: "def",
		Event: "nop",
	}

	err := app.user.GetByUser(u)
	if err != nil {
		log.Println("[DB] ERROR:", err)
	}

	err = app.fsm.SendEvent(EventType(u.Event), actionContext)
	if err != nil {
		log.Println("[FSM] ERROR:", err)
	}

	u.State = string(app.fsm.currentState)
	u.Event = string(app.fsm.currentEvent)

	err = app.user.Update(u)
	if err != nil {
		log.Println("[DB] ERROR:", err)
	}
}

func (app *application) startOperation() {
	u := User{
		ID:    app.telegram.upd.Message.From.ID,
		Name:  app.telegram.upd.Message.From.UserName,
		State: "def",
		Event: "nop",
	}
	err := app.user.Upsert(u)
	if err != nil {
		log.Println("[DB] ERROR:", err)
	}
	app.telegram.startOperationResponse()
}

func (app *application) cancelOperation() {
	u := &User{
		ID:   app.telegram.upd.Message.From.ID,
		Name: app.telegram.upd.Message.From.UserName,
	}
	err := app.user.GetByUser(u)
	if err != nil {
		log.Println("[DB] ERROR:", err)
	}
	oldEvent := u.Event
	u.State = "def"
	u.Event = "nop"
	err = app.user.Update(u)
	if err != nil {
		log.Println("[DB] ERROR:", err)
	}
	app.telegram.cancelOperationResponse(EventType(oldEvent))
}

func (app *application) inlineQueryHandler(upd tgbotapi.Update) {
	userID := upd.InlineQuery.From.ID
	query := upd.InlineQuery.Query
	queryOffset, _ := strconv.Atoi(upd.InlineQuery.Offset)

	stickers, err := app.sticker.GetSticker(userID, query, queryOffset)
	if err != nil {
		log.Println("[Get Stickers] ERROR:", err)
	}

	var results []any

	for _, s := range stickers {
		sticker := tgbotapi.NewInlineQueryResultCachedSticker(s.id[:64], s.id, upd.InlineQuery.Query)
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

	if _, err := app.telegram.bot.Request(inlineConf); err != nil {
		log.Println("[Inline request] ERROR:", err)
	}
}
