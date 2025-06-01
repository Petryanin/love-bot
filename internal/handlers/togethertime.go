// internal/handlers/together.go
package handlers

import (
	"context"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TogetherTimeHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		text := app.Relationship.Duration()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
