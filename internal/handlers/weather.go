package handlers

import (
	"context"
	"log"

	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func WeatherHandler(ws *services.WeatherService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		summary, err := ws.TodaySummary()
		if err != nil {
			log.Printf("failed to get weather summary: %v", err.Error())
			summary = "Не удалось получить погоду 😕"
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   summary,
		})
	}
}
