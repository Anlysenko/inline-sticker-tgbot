package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (tg *TelegramModel) sendMessage(message string) {
	msg := tgbotapi.NewMessage(tg.upd.Message.Chat.ID, message)
	tg.bot.Send(msg)
}

func (tg *TelegramModel) startOperationResponse() {
	msg :=
		"Hi, there! I am able to save stickers for later inline use.\n" +
			"To start adding a sticker, please use the /add command " +
			"or simply send sticker to me.\n" +
			"Type /help for info."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) helpOperationResponse() {
	msg :=
		"You can control me by sending these commands:\n" +
			"/add	 - to add a new sticker or change the old one.\n" +
			"/delete - to remove the sticker.\n" +
			"/show	 - to show the tags of the saved sticker.\n" +
			"/cancel - to cancel the current operation."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) unrecognizedCommandResponse() {
	msg := "Unrecognized command. Type /help for info."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) cancelOperationResponse(event EventType) {
	var msg string
	switch event {
	case nop:
		msg = "I wasn't doing anything."
	case addStickerEvent:
		msg = "The process of sending the sticker has been cancelled."
	case addTagsEvent:
		msg = "The process of sending the sticker tags has been cancelled."
	case showTagsEvent:
		msg = "The process of showing the sticker tags has been canceled."
	case deleteStickerEvent:
		msg = "The process of deleting the sticker has been cancelled."
	}
	tg.sendMessage(msg)
}

func (tg *TelegramModel) defaultStateResponse(cmd string) {
	var msg string
	switch cmd {
	case "add":
		msg = "Send me the sticker you would like to add."
	case "show":
		msg = "Send me the sticker for which you would like to view the tags."
	case "delete":
		msg = "Send me the sticker you would like to delete."
	default:
		tg.unrecognizedCommandResponse()
		return
	}
	tg.sendMessage(msg)
}

func (tg *TelegramModel) notStickerResponse() {
	msg := "Maaan... This is not a sticker. Please try again or type /cancel if you wish to end this task."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) databaseErrorResponse() {
	msg := "I apologize, I am unable to process your sticker. It may be due to a database problem. /cancel"
	tg.sendMessage(msg)
}

func (tg *TelegramModel) notFoundResponse() {
	msg := "I am not familiar with requested resource. Please try again or type /cancel if you wish to end this task."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) addingStickerResponse() {
	msg := "Great! Now send me a short description, 1-4 tags separated by spaces."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) addingTagsResponse(cmd string) {
	var msg string
	switch cmd {
	case "success":
		msg = "Nice! Your sticker has been added."
	case "invalid_tags":
		msg = "This description is unavailable.\n" +
			"The description must contain at least one tag " +
			"and no more than four tags, separated by spaces.\n" +
			"Only use letters, numbers, hyphens and spaces.\n" +
			"Please try again or type /cancel if you wish to end this task."
	case "update_error":
		msg = "You took too long to come up with a description. " +
			"Please resubmit the sticker and try to add the tags within 5 minutes."
	}
	tg.sendMessage(msg)
}

func (tg *TelegramModel) showingTagsResponse(cmd string) {
	msg := fmt.Sprintf("This sticker has the following tags: `%s`\n", cmd)
	msg += "You can continue sending stickers for their tag information " +
		"or cancel the current operation using the /cancel command."
	tg.sendMessage(msg)
}

func (tg *TelegramModel) deletingStickerResponse() {
	msg := "Nice! The sticker has been deleted."
	tg.sendMessage(msg)
}
