package main

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ProcessSrartOperation(bot tgbotapi.BotAPI, upd tgbotapi.Update) {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
	msg.Text = "Hi! I can save stickers for further inline use.\n" +
		"Start adding a sticker by the /add command " +
		"or just send it to me.\n" +
		"Type /help for info."
	bot.Send(msg)
}

func ProcessUnrecognizedCommand(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	if upd.Message.Sticker == nil {
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
		msg.Text = "Unrecognized command. Type /help for info."
		bot.Send(msg)
		return
	}
	redis.State = addingSticker
	redis.Event = sendToAddSticker
	ProcessFSM(bot, upd, redis)
}

func ProcessHelpOperation(bot tgbotapi.BotAPI, upd tgbotapi.Update) {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
	msg.Text = "You can control me by sending these commands:\n" +
		"/add	 - to add a new sticker or change the old one.\n" +
		"/delete - to delete the sticker.\n" +
		"/show	 - to show the description of the saved sticker.\n" +
		"/cancel - to cancel current operation."
	bot.Send(msg)
}

func ProcessCancelOpearation(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
	switch redis.State {
	case Def:
		msg.Text = "I wasn't doing anything."
	case addingSticker:
		msg.Text = "Sending sticker operation has been canceled."
	case sendingToAddSticker, sendingToAddDescription:
		msg.Text = "Sending sticker description operation has been canceled."
		err := DeleteStickerPG(upd.Message.From.ID, "~")
		if err != nil {
			log.Println("[Delete sticker postgres] ERROR:", err)
		}
	case deletingSticker, sendingToDeleteSticker:
		msg.Text = "Deleting sticker operation has been canceled."
	case showingDescription, sendingToShowDescription:
		msg.Text = "Showing sticker description has been canceled."
	}

	redis.State = Def
	redis.Event = Nop
	err := UpdateDataRedis(upd.Message.From.UserName, upd.Message.Chat.ID, redis)
	if err != nil {
		log.Println("[Update state redis] ERROR:", err)
	}

	bot.Send(msg)
}

func InitAddingSticker(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	redis.State = Def
	redis.Event = addSticker
	ProcessFSM(bot, upd, redis)
}

func InitDeletingSticker(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	redis.State = Def
	redis.Event = deleteSticker
	ProcessFSM(bot, upd, redis)
}

func InitShowDescription(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	redis.State = Def
	redis.Event = showDescription
	ProcessFSM(bot, upd, redis)
}

func ProcessFSM(bot tgbotapi.BotAPI, upd tgbotapi.Update, redis *uidChidRedis) {
	stickerFSM := processStickerFSM(redis.State)
	addingCtx := &ProcessingStickerContext{
		bot:   bot,
		upd:   upd,
		redis: redis,
	}

	state, event, err := stickerFSM.SendEvent(redis.Event, addingCtx)
	if err != nil {
		log.Println("[FSM] ERROR:", err)
	}

	redis = &uidChidRedis{Event: event, State: state}
	err = UpdateDataRedis(upd.Message.From.UserName, upd.Message.Chat.ID, redis)
	if err != nil {
		log.Println("[Update state redis] ERROR:", err)
	}
}

type ProcessingStickerContext struct {
	bot   tgbotapi.BotAPI
	upd   tgbotapi.Update
	redis *uidChidRedis
}

type addingStickerAction struct{}

func (a *addingStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "Send me the sticker you want to add.")
	ctx.bot.Send(msg)

	return sendToAddSticker
}

type sendingStickerAction struct{}

func (a *sendingStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)
	userID := ctx.upd.Message.From.ID
	sticker := ctx.upd.Message.Sticker

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "")

	if sticker == nil {
		msg.Text = "Maaan... This is not a sticker. Try again or /cancel"
		ctx.bot.Send(msg)
		return sendToAddSticker
	}

	msg.Text = "Great! Now send me a brief description within 4 words."
	if err := InsertStickerPG(userID, sticker.FileID, sticker.FileUniqueID, "~"); err != nil {
		log.Println("[Insert sticker postgres] ERROR:", err)
		msg.Text = "Sorry, I can't save your sticker. Probably it's DB problem. /cancel"
		ctx.bot.Send(msg)
		return Nop
	}

	ctx.bot.Send(msg)
	return sendToAddDescription
}

type sendingDescriptionAction struct{}

func (a *sendingDescriptionAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)
	userID := ctx.upd.Message.From.ID
	desc := ctx.upd.Message

	msg := tgbotapi.NewMessage(desc.Chat.ID, "")

	if !isDescriptionAvailable(desc.Text) {
		msg.Text = "This description is not available.\n" +
			"You can write from 1 to 4 words separated by spaces. " +
			"Use only letters, numbers, dashes and spaces.\n" +
			"Try again or /cancel"
		ctx.bot.Send(msg)
		return sendToAddDescription
	}

	msg.Text = "Nice! Sticker has been added."
	if err := UpdateDescriptionPG(userID, desc.Text); err != nil {
		log.Println("[Update description postgres] ERROR:", err)
		msg.Text = "Sorry, I can't save your sticker. Probably it's DB problem. /cancel"
		ctx.bot.Send(msg)
		return Nop
	}

	ctx.bot.Send(msg)
	return Nop
}

func isDescriptionAvailable(txt string) bool {
	if len(txt) < 2 {
		return false
	}
	f := func(c rune) bool {
		isSpace := c == rune(' ')
		isHyphen := c == rune('-')
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !isSpace && !isHyphen
	}
	for _, r := range txt {
		if f(r) {
			return false
		}
	}

	return len(strings.Fields(txt)) <= 4
}

type deletingStickerAction struct{}

func (a *deletingStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "Send me the sticker you want to delete.")
	ctx.bot.Send(msg)

	return sendToDeleteSticker
}

type sendingToDeleteStickerAction struct{}

func (a *sendingToDeleteStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)
	userID := ctx.upd.Message.From.ID
	sticker := ctx.upd.Message.Sticker

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "")

	if sticker == nil {
		msg.Text = "Maaan... This is not a sticker. Try again or /cancel"
		ctx.bot.Send(msg)
		return sendToDeleteSticker
	}

	msg.Text = "Nice! Sticker has been deleted."
	if err := DeleteStickerPG(userID, sticker.FileUniqueID); err != nil {
		log.Println("[Delete sticker postgres] ERROR:", err)
		msg.Text = "Sorry, I can't delete your sticker. Probably it's DB problem. /cancel"
		ctx.bot.Send(msg)
		return Nop
	}

	ctx.bot.Send(msg)
	return Nop
}

type showingDescriptionAction struct{}

func (a *showingDescriptionAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "Send me the sticker whose description you want to see.")
	ctx.bot.Send(msg)

	return sendToShowDescription
}

type sendingToShowDescriptionAction struct{}

func (a *sendingToShowDescriptionAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ProcessingStickerContext)
	userID := ctx.upd.Message.From.ID
	sticker := ctx.upd.Message.Sticker

	msg := tgbotapi.NewMessage(ctx.upd.Message.Chat.ID, "")

	if sticker == nil {
		msg.Text = "Maaan... This is not a sticker. Try again or /cancel"
		ctx.bot.Send(msg)
		return sendToShowDescription
	}

	desc, err := GetStickerDescriptionPG(userID, sticker.FileUniqueID)
	if err != nil {
		log.Println("[Show description postgres] ERROR:", err)
		msg.Text = "Sorry, I can't show your description. Probably it's DB problem. /cancel"
		ctx.bot.Send(msg)
		return Nop
	}

	if desc == "" {
		msg.Text = "I know nothing about this sticker. Try again or /cancel"
		ctx.bot.Send(msg)
		return sendToShowDescription
	}

	msg.Text = fmt.Sprintf("Description of this sticker: %s", desc)
	ctx.bot.Send(msg)
	return Nop
}
