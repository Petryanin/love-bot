// internal/handlers/together.go
package handlers

import (
	"context"

	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TogetherTimeHandler(rs *services.RelationshipService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		text := rs.Duration()

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
