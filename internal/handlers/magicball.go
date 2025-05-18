package handlers

import (
	"context"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func MagicBallHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   appCtx.MagicBallService.Answer(),
		})
	}
}
