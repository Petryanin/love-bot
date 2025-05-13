package handlers

import (
	"context"

	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DefaultReplyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Нажми кнопку или введи /help, чтобы узнать мои возможности🤗",
		ReplyMarkup: keyboards.BaseReplyKeyboard(),
	})
}

func FallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Извини, я тебя не понял 😿\nНажми кнопку или введи /help, чтобы узнать мои возможности🤗",
		ReplyMarkup: keyboards.BaseReplyKeyboard(),
	})
}
