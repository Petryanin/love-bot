package handlers

import (
	"context"
	"log"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func WeatherHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		user, err := app.User.Get(ctx, db.WithChatID(chatID))
		if err != nil {
			// todo исправить логику
			log.Print("handlers: failed to get user info: %w", err)
			app.Session.Reset(chatID)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Упс, не удалось получить погоду 😿\nПопробуй позже",
			})
			return
		}

		summary, err := app.Weather.TodaySummary(ctx, user.City)
		if err != nil {
			log.Printf("handlers: failed to get weather summary: %v", err.Error())
			summary = "Не удалось получить погоду 😕"
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   summary,
		})
	}
}
