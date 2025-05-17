package handlers

import (
	"context"
	"fmt"
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
		fmt.Sprintf("\"%s\"", config.WeatherBtn) + " — краткая сводка текущей погоды",
		fmt.Sprintf("\"%s\"", config.TogetherTimeBtn) + " — время ваших отношений",
		fmt.Sprintf("\"%s\"", strings.Replace(config.ComplimentBtn, "-", "\\-", -1)) + " — картинка с котом и комплиментом",
	}, "\n")
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      helpText,
		ParseMode: models.ParseModeMarkdown,
	})
}
