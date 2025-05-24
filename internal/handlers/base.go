package handlers

import (
	"context"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DefaultReplyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Нажми кнопку или введи /help, чтобы узнать мои возможности🤗",
		ReplyMarkup: keyboards.BaseReplyKeyboard(),
	})
}

func FallbackHandler(kb *models.ReplyKeyboardMarkup) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		chatID := update.Message.Chat.ID

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Извини, я тебя не понял 😿\nНажми кнопку или введи /help, чтобы узнать мои возможности🤗",
			ReplyMarkup: kb,
		})
	}
}

func StateRootHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)
		text := upd.Message.Text

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		if sess.State == services.StateRoot {
			switch text {
			case config.PlansBtn:
				PlansHandler(appCtx)(ctx, b, upd)
			case config.SettingsBtn:
				SettingsHandler(appCtx)(ctx, b, upd)
			default:
				FallbackHandler(keyboards.BaseReplyKeyboard())(ctx, b, upd)
			}
			return
		}

		switch {
		case appCtx.SessionManager.IsPlanState(chatID):
			PlansHandler(appCtx)(ctx, b, upd)
		case appCtx.SessionManager.IsSettingsState(chatID):
			SettingsHandler(appCtx)(ctx, b, upd)
		}
	}
}
