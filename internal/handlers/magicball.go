package handlers

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func MagicBallHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		file, err := app.MagicBall.ImageAnswer()
		if err != nil {
			log.Print("handlers: failed to get magic ball image file: %w", err)
		}

		emulateThinkingProcess(ctx, b, update)

		b.SendSticker(ctx, &bot.SendStickerParams{
			ChatID: chatID,
			Sticker: &models.InputFileUpload{
				Data:     file,
				Filename: "sticker.webp",
			},
		})
	}
}

func emulateThinkingProcess(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	msg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "🔮",
	})
	time.Sleep(1300 * time.Millisecond)

	animations := []func(){
		func() { animateDots(ctx, b, chatID, msg.ID) },
		func() { animateProgressBar(ctx, b, chatID, msg.ID) },
		func() { animateSymbols(ctx, b, chatID, msg.ID) },
	}

	animations[rand.IntN(len(animations))]()

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: msg.ID,
	})
}

func animateDots(ctx context.Context, b *bot.Bot, chatID int64, msgID int) {
	texts := []string{
		"🔮 Я вижу тьму",
		"🔮 Образы начинают проясняться",
		"🔮 Судьба медленно раскрывается",
	}

	for _, phrase := range texts {
		for dots := 0; dots <= 3; dots++ {
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: msgID,
				Text:      phrase + strings.Repeat(".", dots),
			})
			time.Sleep(600 * time.Millisecond)
		}
	}
}

func animateProgressBar(ctx context.Context, b *bot.Bot, chatID int64, msgID int) {
	type barStyle struct {
		filled  string
		empty   string
		caption string
	}

	styles := []barStyle{
		{"▓", "░", "🔮 Зарядка шара"},
		{"■", "□", "🔮 Энергия накапливается"},
		{"🟦", "⬜", "🔮 Оракул пробуждается"},
		{"✶", "⋅", "🔮 Пророчество формируется"},
	}

	style := styles[rand.IntN(len(styles))]

	width := 12
	for i := 1; i <= width; i++ {
		bar := strings.Repeat(style.filled, i) + strings.Repeat(style.empty, width-i)
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: msgID,
			Text:      fmt.Sprintf("%s...\n[%s]", style.caption, bar),
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func animateSymbols(ctx context.Context, b *bot.Bot, chatID int64, msgID int) {
	patterns := []string{
		"⎈ ☯ ☸ ☥ ⚛ ♾ ✴",
		"⚛ ☽ ☼ ✪ ☄ ⚡ ❂",
		"🜁 🜂 🜃 🜄 🝔 🜚 🝎",
		"𓂀 𓃠 𓆑 𓏏 𓄿 𓅓 𓊃",
		"✵ ✸ ✺ ✹ ✷ ✶ ✳",
		"⌘ ⚙ ⚒ ⚔ ⛧ ⚚ ☸",
		"◐ ◓ ◑ ◒ ◐ ◓ ◑ ◒",
		"𝕄𝕒𝕘𝕚𝕔 𝕚𝕤 𝕨𝕠𝕣𝕜𝕚𝕟𝕘...",
	}
	for _, p := range patterns {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: msgID,
			Text:      fmt.Sprintf("🔮 %s", p),
		})
		time.Sleep(700 * time.Millisecond)
	}
}
