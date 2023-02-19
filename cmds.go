package main

import (
	"errors"
	"log"
	"strings"
	"unicode"
)

type ActionContext struct {
	tgModel    *TelegramModel
	sticker    StickerModel
	user       UserModel
	memSticker *MemSticker
}

type defaultStickerAction struct{}

func (a *defaultStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ActionContext)

	sticker := ctx.tgModel.upd.Message.Sticker
	userID := ctx.tgModel.upd.Message.From.ID
	if sticker != nil {
		ctx.memSticker.addSticker(userID, sticker)
		ctx.tgModel.addingStickerResponse()
		return addTagsEvent
	}

	requestedCommand := ctx.tgModel.upd.Message.Command()
	ctx.tgModel.defaultStateResponse(requestedCommand)

	switch requestedCommand {
	case "add":
		return addStickerEvent
	case "show":
		return showTagsEvent
	case "delete":
		return deleteStickerEvent
	default:
		return nop
	}
}

type addingStickerAction struct{}

func (a *addingStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ActionContext)

	sticker := ctx.tgModel.upd.Message.Sticker
	userID := ctx.tgModel.upd.Message.From.ID

	if sticker == nil {
		ctx.tgModel.notStickerResponse()
		return addStickerEvent
	}

	ctx.memSticker.addSticker(userID, sticker)
	ctx.tgModel.addingStickerResponse()
	return addTagsEvent
}

type addingTagsAction struct{}

func (a *addingTagsAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ActionContext)

	userID := ctx.tgModel.upd.Message.From.ID
	tags := ctx.tgModel.upd.Message.Text

	if !a.isTagsAvailable(tags) {
		ctx.tgModel.addingTagsResponse("invalid_tags")
		return addTagsEvent
	}

	if err := ctx.memSticker.updateSticker(userID, tags); err != nil {
		ctx.tgModel.addingTagsResponse("update_error")
		return nop
	}

	if err := ctx.sticker.InsertSticker(ctx.memSticker.userStickersMap[userID].sticker); err != nil {
		log.Println("[Update tags postgres] ERROR:", err)
		ctx.tgModel.databaseErrorResponse()
		return addTagsEvent
	}

	ctx.tgModel.addingTagsResponse("success")
	return nop
}

func (a *addingTagsAction) isTagsAvailable(input string) bool {
	if len(input) > 100 {
		return false
	}

	words := strings.Fields(input)
	if len(words) < 1 || len(words) > 4 {
		return false
	}

	for _, word := range words {
		for _, r := range word {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' {
				return false
			}
		}
	}
	return true
}

type showingTagsAction struct{}

func (a *showingTagsAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ActionContext)

	userID := ctx.tgModel.upd.Message.From.ID
	sticker := ctx.tgModel.upd.Message.Sticker

	if sticker == nil {
		ctx.tgModel.notStickerResponse()
		return showTagsEvent
	}

	tags, err := ctx.sticker.GetStickerTags(userID, sticker.FileUniqueID)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			ctx.tgModel.notFoundResponse()
			return showTagsEvent
		default:
			log.Println("[Show tags postgres] ERROR:", err)
			ctx.tgModel.databaseErrorResponse()
			return showTagsEvent
		}
	}

	ctx.tgModel.showingTagsResponse(tags)
	return showTagsEvent
}

type deletingStickerAction struct{}

func (a *deletingStickerAction) Execute(eventCtx EventContext) EventType {
	ctx := eventCtx.(*ActionContext)

	userID := ctx.tgModel.upd.Message.From.ID
	sticker := ctx.tgModel.upd.Message.Sticker

	if sticker == nil {
		ctx.tgModel.notStickerResponse()
		return deleteStickerEvent
	}

	if err := ctx.sticker.DeleteSticker(userID, sticker.FileUniqueID); err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			ctx.tgModel.notFoundResponse()
			return deleteStickerEvent
		default:
			log.Println("[Delete sticker postgres] ERROR:", err)
			ctx.tgModel.databaseErrorResponse()
			return deleteStickerEvent
		}
	}

	ctx.tgModel.deletingStickerResponse()
	return nop
}
