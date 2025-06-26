package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func SettingsHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)
		text := upd.Message.Text

		allowedMap := map[string]bool{
			config.CityBtn: true, config.PartnerBtn: true, config.CatBtn: true,
			config.DisableBtn: true, config.BackBtn: true, config.CancelBtn: true,
		}

		if sess.State == services.StateSettingsMenu && !allowedMap[text] {
			FallbackHandler(keyboards.SettingsMenuKeyboard())(ctx, b, upd)
			return
		}

		switch sess.State {
		case services.StateRoot:
			sess.State = services.StateSettingsMenu
			fallthrough
		case services.StateSettingsMenu:
			settingsMenuHandler(app)(ctx, b, upd)
		case services.StateSettingsCity:
			settingsCityHandler(app)(ctx, b, upd)
		case services.StateSettingsPartner:
			settingsPartnerHandler(app)(ctx, b, upd)
		case services.StateSettingsCat:
			settingsCatHandler(app)(ctx, b, upd)

		default:
			FallbackHandler(keyboards.SettingsMenuKeyboard())(ctx, b, upd)
		}
	}
}

func settingsMenuHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		switch text {
		case config.SettingsBtn, config.CancelBtn:
			showCurrentSettings(ctx, b, app, chatID)
		case config.CityBtn:
			sess.State = services.StateSettingsCity
			promptForCity(ctx, b, chatID)
		case config.PartnerBtn:
			sess.State = services.StateSettingsPartner
			promptForPartner(ctx, b, chatID)
		case config.CatBtn:
			sess.State = services.StateSettingsCat
			promptForCatTime(ctx, b, chatID)
		case config.BackBtn:
			app.Session.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)
		}
	}
}

func settingsCityHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		if handleCancel(upd.Message.Text, app, chatID, func() { SettingsHandler(app)(ctx, b, upd) }) {
			return
		}

		city, tz, err := getGeoData(ctx, app, upd)
		if err != nil {
			log.Printf("handlers: failed to get geo data: %v", err)
			sendError(ctx, b, chatID, MsgLocationNotFound)
			return
		}

		app.User.UpdateGeo(ctx, chatID, city, tz)
		app.Session.Get(chatID).State = services.StateSettingsMenu

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf(MsgCitySaved, city, tz),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func settingsPartnerHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text

		if handleCancel(text, app, chatID, func() { SettingsHandler(app)(ctx, b, upd) }) {
			return
		}

		partnerName, err := validateTgUsername(text)
		if err != nil {
			log.Printf("handlers: invalid tg username: %v", err)
			sendError(ctx, b, chatID, MsgInvalidUsername)
			return
		}

		if err := app.User.UpdatePartner(ctx, chatID, partnerName); err != nil {
			log.Printf("handlers: failed to update partner: %v", err)
			sendError(ctx, b, chatID, MsgPartnerNotFound)
			return
		}

		if !strings.HasPrefix(text, "@") {
			text = "@" + text
		}

		app.Session.Get(chatID).State = services.StateSettingsMenu
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf(MsgPartnerSaved, text),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
		})
	}
}

func settingsCatHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text

		if handleCancel(text, app, chatID, func() { SettingsHandler(app)(ctx, b, upd) }) {
			return
		}

		if text == config.DisableBtn {
			disableCatNotifications(ctx, b, app, chatID)
			return
		}

		updateCatTime(ctx, b, app, chatID, text)
	}
}

func showCurrentSettings(ctx context.Context, b *bot.Bot, app *app.App, chatID int64) {
	user, err := app.User.Get(ctx, db.WithChatID(chatID), db.WithPartnerInfo())
	if err != nil {
		log.Printf("handlers: failed to get user info: %v", err)
		app.Session.Reset(chatID)
		sendError(ctx, b, chatID, MsgUserInfoError)
		return
	}

	catTimeStr := MsgCatTimeDisabled
	if !user.CatTime.IsZero() {
		catTimeStr = fmt.Sprintf(MsgCatTimeEnabled, user.CatTime.Format("15:04"))
	}

	msg := fmt.Sprintf(MsgCurrentSettings, user.City, user.TZ.String(), bot.EscapeMarkdown(user.PartnerName), catTimeStr)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        msg,
		ReplyMarkup: keyboards.SettingsMenuKeyboard(),
		ParseMode:   models.ParseModeMarkdown,
	})
}

func promptForCity(ctx context.Context, b *bot.Bot, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        MsgCityPrompt,
		ReplyMarkup: keyboards.CancelKeyboard(),
	})
}

func promptForPartner(ctx context.Context, b *bot.Bot, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        MsgPartnerPromptSettings,
		ReplyMarkup: keyboards.CancelKeyboard(),
	})
}

func promptForCatTime(ctx context.Context, b *bot.Bot, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        fmt.Sprintf(MsgCatPrompt, config.DisableBtn),
		ReplyMarkup: keyboards.DisableKeyboard(),
	})
}

func handleCancel(text string, app *app.App, chatID int64, cancelFunc func()) bool {
	if text == config.CancelBtn {
		app.Session.Get(chatID).State = services.StateSettingsMenu
		cancelFunc()
		return true
	}
	return false
}

func disableCatNotifications(ctx context.Context, b *bot.Bot, app *app.App, chatID int64) {
	if err := app.User.UpdateCatTime(ctx, chatID, ""); err != nil {
		log.Printf("handlers: failed to update cat time: %v", err)
		sendError(ctx, b, chatID, MsgCatTimeUpdateError)
		return
	}
	app.Session.Get(chatID).State = services.StateSettingsMenu
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        MsgCatTimeDisabledSuccess,
		ReplyMarkup: keyboards.SettingsMenuKeyboard(),
	})
}

func updateCatTime(ctx context.Context, b *bot.Bot, app *app.App, chatID int64, text string) {
	user, err := app.User.Get(ctx, db.WithChatID(chatID))
	if err != nil {
		log.Printf("handlers: failed to get user info: %v", err)
		sendError(ctx, b, chatID, MsgUserInfoError)
		return
	}

	if _, err := time.ParseInLocation("15:04", text, user.TZ); err != nil {
		sendError(ctx, b, chatID, MsgInvalidTimeFormat)
		return
	}

	if err := app.User.UpdateCatTime(ctx, chatID, text); err != nil {
		log.Printf("handlers: failed to update cat time: %v", err)
		sendError(ctx, b, chatID, MsgCatTimeSaveError)
		return
	}

	app.Session.Get(chatID).State = services.StateSettingsMenu
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        fmt.Sprintf(MsgCatTimeSaved, text),
		ReplyMarkup: keyboards.SettingsMenuKeyboard(),
	})
}
