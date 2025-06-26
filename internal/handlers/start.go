package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)

		if sess.State != services.StateRoot && !app.Session.IsStartSettingsState(chatID) {
			sendError(ctx, b, chatID, MsgCommandNotAvailable)
			return
		}

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		switch sess.State {
		case services.StateRoot:
			app.Session.Reset(chatID)
			_, err := app.User.Get(ctx, db.WithChatID(chatID))
			if err != nil {
				StartSettingsHandler(app)(ctx, b, upd)
				return
			}
			sendWelcomeMessage(ctx, b, chatID)

		case services.StateStartCity:
			startCityHandler(app)(ctx, b, upd)

		case services.StateStartPartner:
			startPartnerHandler(app)(ctx, b, upd)
		}
	}
}

func StartSettingsHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)
		sess.State = services.StateStartCity

		text := MsgWelcome + bot.EscapeMarkdown(MsgSettingsStart)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}

func startCityHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID

		city, tz, err := getGeoData(ctx, app, upd)
		if err != nil {
			log.Printf("handlers: failed to get geo data: %v", err)
			sendError(ctx, b, chatID, MsgLocationNotFound)
			return
		}

		sess := app.Session.Get(chatID)
		sess.TempStartCity = city
		sess.TempStartTZ = tz
		sess.State = services.StateStartPartner

		msg := fmt.Sprintf(MsgCitySaved, city, tz) + bot.EscapeMarkdown(MsgPartnerPrompt)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      msg,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}

func startPartnerHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		partnerName, err := validateTgUsername(text)
		if err != nil {
			log.Printf("handlers: invalid tg username: %v", err)
			sendError(ctx, b, chatID, MsgInvalidUsername)
			return
		}

		partner, err := app.User.Get(ctx, db.WithUsername(partnerName))
		if err != nil {
			log.Printf("handlers: failed to get partner: %v", err)
			sendError(ctx, b, chatID, MsgPartnerNotFound)
			return
		}

		err = app.User.Upsert(ctx, chatID, upd.Message.From.Username, sess.TempStartCity, sess.TempStartTZ, partner.ChatID)
		if err != nil {
			log.Printf("handlers: failed to upsert user: %v", err)
			sendError(ctx, b, chatID, MsgPartnerNotFound)
			return
		}

		app.Session.Reset(chatID)

		if !strings.HasPrefix(text, "@") {
			text = "@" + text
		}

		msg := bot.EscapeMarkdown(fmt.Sprintf(MsgPartnerSaved, text)+MsgSettingsComplete+MsgStartBase+MsgStartDeveloper) +
			MsgStartSource + MsgStartGratitude

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: keyboards.BaseReplyKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func sendWelcomeMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	text := MsgWelcome + bot.EscapeMarkdown(MsgStartBase+MsgStartDeveloper) + MsgStartSource + MsgStartGratitude
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboards.BaseReplyKeyboard(),
		ParseMode:   models.ParseModeMarkdown,
	})
}
