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

const (
	dotsAnimationSpeed         = 600 * time.Millisecond
	progressBarAnimationSpeed  = 500 * time.Millisecond
	symbolsAnimationSpeed      = 700 * time.Millisecond
	thinkingProcessInitialWait = 1300 * time.Millisecond
	progressBarWidth           = 12
)

func MagicBallHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID := update.Message.Chat.ID

		file, err := app.MagicBall.ImageAnswer()
		if err != nil {
			log.Printf("handlers: failed to get magic ball image file: %v", err)
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

	time.Sleep(thinkingProcessInitialWait)

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
		MsgMagicBallThinking1,
		MsgMagicBallThinking2,
		MsgMagicBallThinking3,
	}

	for _, phrase := range texts {
		for dots := 0; dots <= 3; dots++ {
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: msgID,
				Text:      phrase + strings.Repeat(".", dots),
			})
			time.Sleep(dotsAnimationSpeed)
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
		{"▓", "░", MsgMagicBallCharging},
		{"■", "□", MsgMagicBallAccumulating},
		{"🟦", "⬜", MsgMagicBallAwakening},
		{"✶", "⋅", MsgMagicBallProphesying},
	}

	style := styles[rand.IntN(len(styles))]

	for i := 1; i <= progressBarWidth; i++ {
		bar := strings.Repeat(style.filled, i) + strings.Repeat(style.empty, progressBarWidth-i)
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: msgID,
			Text:      fmt.Sprintf("%s...\n[%s]", style.caption, bar),
		})
		time.Sleep(progressBarAnimationSpeed)
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
		time.Sleep(symbolsAnimationSpeed)
	}
}
