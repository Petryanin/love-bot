package handlers

import (
	"context"
	"strings"

	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	kb := keyboards.BaseReplyKeyboard()

	welcomeText := strings.Join([]string{
		"Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n",
		"Можешь сразу нажать на одну из кнопок или позвать команду /help, ",
		"чтобы ознакомиться с моими функциями подробнее",
	}, "")
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text: welcomeText,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: kb,
	})
}
