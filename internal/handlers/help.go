package handlers

import (
	"context"
	"strings"

	"github.com/Petryanin/love-bot/internal/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func HelpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	helpText := strings.Join([]string{
		"Вот что я умею:",
		"",
		"*Команды:*",
		"/start — приветствие и главное меню",
		"/help — показать это сообщение",
		"",
		"*Кнопки:*",
		"\"" + config.WeatherBtn + "\" — краткая сводка текущей погоды",
		"\"" + strings.Replace(config.ComplimentBtn, "-", "\\-", -1) + "\" — картинка с котом и комплиментом",
		"\"" + config.PlansBtn + "\" — меню планов и напоминаний",
		"\"" + config.TogetherTimeBtn + "\" — время ваших отношений",
		"\"" + config.MagicBallBtn + "\" — поможет тебе принять решение",
	}, "\n")
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      helpText,
		ParseMode: models.ParseModeMarkdown,
	})
}
