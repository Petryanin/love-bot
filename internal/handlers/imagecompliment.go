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

// sendComplimentImage отправляет комплимент в виде изображения или текста, если генерация не удалась
func sendComplimentImage(ctx context.Context, b *bot.Bot, app *app.App, chatID int64) {
	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionUploadPhoto,
	})

	compliment := app.Compliment.Random()

	imgBytes, err := app.ImageCompliment.Generate(ctx, compliment)
	if err != nil {
		log.Printf("failed to generate compliment image: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   MsgComplimentImageError,
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

// ComplimentImageHandler обрабатывает запрос на получение комплимента от пользователя
func ComplimentImageHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		sendComplimentImage(ctx, b, app, update.Message.Chat.ID)
	}
}

// ScheduledComplimentImageHandler обрабатывает отправку комплимента по расписанию
func ScheduledComplimentImageHandler(ctx context.Context, app *app.App, b *bot.Bot, chatID int64) {
	sendComplimentImage(ctx, b, app, chatID)
}
