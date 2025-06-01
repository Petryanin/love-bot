package handlers

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func ComplimentImageHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionUploadPhoto,
		})

		compliment := app.Compliment.Random()

		imgBytes, err := app.ImageCompliment.Generate(ctx, compliment)
		if err != nil {
			log.Print(err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Упс, комплимент где-то зажевался😿. Пишу что-то от себя...",
			})
			b.SendChatAction(ctx, &bot.SendChatActionParams{
				ChatID: chatID,
				Action: models.ChatActionTyping,
			})
			time.Sleep(300 * time.Millisecond)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   compliment,
			})
			return
		}

		b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: chatID,
			Photo: &models.InputFileUpload{
				Filename: "cat.png",
				Data:     bytes.NewReader(imgBytes),
			},
		})
	}
}

func ScheduledComplimentImageHandler(ctx context.Context, app *app.App, b *bot.Bot, chatID int64) {
	update := &models.Update{Message: &models.Message{Chat: models.Chat{ID: chatID}}}
	ComplimentImageHandler(app)(ctx, b, update)
}
