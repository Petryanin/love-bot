package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartHandler(codeLines int) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		kb := keyboards.BaseReplyKeyboard()
		linesCount := fmt.Sprintf("%d %s", codeLines, services.Pluralize(codeLines, "строка", "строки", "строк"))

		welcomeText := strings.Join([]string{
			"Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n",
			bot.EscapeMarkdown("Можешь сразу нажать на одну из кнопок или позвать команду /help, "),
			bot.EscapeMarkdown("чтобы ознакомиться с моими функциями подробнее.\n\n\n"),
			bot.EscapeMarkdown("От разработчика:\n"),
			bot.EscapeMarkdown("Меня зовут Алексей @Petryanin\n"),
			fmt.Sprintf("Исходный код открыт — на данный момент проект насчитывает *%s*\\.\n", linesCount),
			"Я очень старался и буду благодарен за ⭐ на [github](https://github.com/Petryanin/love-bot)💖",
		}, "")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        welcomeText,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: kb,
		})
	}
}
