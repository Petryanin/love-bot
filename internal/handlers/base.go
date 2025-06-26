package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const telegramUsernameRegex = `^[a-zA-Z][a-zA-Z0-9_]{4,31}$`

func validateTgUsername(input string) (string, error) {
	username := input
	if len(username) > 0 && username[0] == '@' {
		username = username[1:]
	}

	match, _ := regexp.MatchString(telegramUsernameRegex, username)
	if !match {
		return "", fmt.Errorf("wrong telegram username format: %s", username)
	}

	return username, nil
}

func getGeoData(ctx context.Context, app *app.App, upd *models.Update) (string, string, error) {
	if upd.Message.Text != "" {
		return app.Geo.ResolveByName(ctx, upd.Message.Text)
	} else if upd.Message.Location != nil {
		return app.Geo.ResolveByCoords(ctx, upd.Message.Location.Latitude, upd.Message.Location.Longitude)
	}
	return "", "", fmt.Errorf(MsgUnknownMessageType)
}

func sendError(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}

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

func StateRootHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		if strings.HasPrefix(upd.Message.Text, "/") {
			return
		}

		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)
		text := upd.Message.Text

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		if sess.State == services.StateRoot {
			switch text {
			case config.PlansBtn:
				PlansHandler(app)(ctx, b, upd)
			case config.SettingsBtn:
				SettingsHandler(app)(ctx, b, upd)
			default:
				FallbackHandler(keyboards.BaseReplyKeyboard())(ctx, b, upd)
			}
			return
		}

		switch {
		case app.Session.IsPlanState(chatID):
			PlansHandler(app)(ctx, b, upd)
		case app.Session.IsSettingsState(chatID):
			SettingsHandler(app)(ctx, b, upd)
		case app.Session.IsStartSettingsState(chatID):
			StartHandler(app)(ctx, b, upd)
		}
	}
}
