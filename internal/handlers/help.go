package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func HelpHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)

		if sess.State != services.StateRoot {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   MsgCommandNotAvailable,
			})
			return
		}
		app.Session.Reset(chatID)

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: upd.Message.Chat.ID,
			Text: fmt.Sprintf(
				MsgHelp,
				config.WeatherBtn,
				strings.ReplaceAll(config.ComplimentBtn, "-", "\\-"),
				config.PlansBtn,
				config.TogetherTimeBtn,
				config.MagicBallBtn,
				config.SettingsBtn,
			),
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
