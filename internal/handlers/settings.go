package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"

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

func SettingsHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)
		text := upd.Message.Text

		var allowedMap = map[string]bool{
			config.CityBtn:    true,
			config.PartnerBtn: true,
			config.BackBtn:    true,
			config.CancelBtn:  true,
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
			settingsMenuHandler(appCtx)(ctx, b, upd)
			return

		case services.StateSettingsCity:
			settingsCityHandler(appCtx)(ctx, b, upd)
			return

		case services.StateSettingsPartner:
			settingsPartnerHandler(appCtx)(ctx, b, upd)
			return
		}

		FallbackHandler(keyboards.SettingsMenuKeyboard())(ctx, b, upd)
	}
}

func settingsMenuHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		switch {
		case text == config.SettingsBtn || text == config.CancelBtn:
			user, err := appCtx.UserService.GetByID(ctx, chatID, false)
			if err != nil {
				log.Print("handlers: failed to get user info: %w", err)
				appCtx.SessionManager.Reset(chatID)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:      chatID,
					Text:        "Упс, не удалось получить информацию по пользователю 😿\nПопробуй позже",
					ReplyMarkup: keyboards.BaseReplyKeyboard(),
				})
				return
			}

			tz := user.TZ.String()
			text = fmt.Sprintf(
				"*Ваши текущие настройки:*\n\n"+
					"\\- город: *%s*\n"+
					"\\- часовой пояс: *%s* \n"+
					"\\- партнер: @%s",
				user.City, tz, bot.EscapeMarkdown(user.PartnerName),
			)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        text,
				ReplyMarkup: keyboards.SettingsMenuKeyboard(),
				ParseMode:   models.ParseModeMarkdown,
			})

		case text == config.CityBtn:
			sess.State = services.StateSettingsCity
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Пожалуйста, отправьте мне название своего города или вашу геолокацию.\n\n" +
					"Это поможет мне давать вам актуальную сводку погоды и учитывать часовой пояс при напоминаниях",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

		case text == config.PartnerBtn:
			sess.State = services.StateSettingsPartner
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Пожалуйста, введите Telegram-ник вашего партнёра.\n\n" +
					"Это поможет мне учитывать ваши совместные планы.",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

		case text == config.BackBtn:
			appCtx.SessionManager.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)
		}
	}
}

func settingsCityHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		var city string
		var tz string
		var err error

		if text != "" {
			if text == config.CancelBtn {
				sess.State = services.StateSettingsMenu
				SettingsHandler(appCtx)(ctx, b, upd)
				return
			}

			city, tz, err = appCtx.GeoService.ResolveByName(ctx, text)
			if err != nil {
				log.Print("handlers: failed to resolve geo info by city name: %w", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать город, попробуй ещё",
				})
				return
			}
		} else if upd.Message.Location != nil {
			city, tz, err = appCtx.GeoService.ResolveByCoords(ctx, upd.Message.Location.Latitude, upd.Message.Location.Longitude)
			if err != nil {
				log.Print("handlers: failed to resolve geo info by coords: %w", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать геолокацию, попробуй ещё",
				})
				return
			}
		}

		appCtx.UserService.UpdateGeo(ctx, chatID, city, tz)
		sess.State = services.StateSettingsMenu

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf("Город сохранён: *%s*\nЧасовой пояс: *%s*", city, tz),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func settingsPartnerHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StateSettingsMenu
			SettingsHandler(appCtx)(ctx, b, upd)
			return
		}

		partnerName, err := validateTgUsername(text)
		if err != nil {
			log.Print("handlers: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "🧐Не смог распознать имя пользователя." +
					"Проверь правильность написания и попробуй ещё",
			})
			return
		}

		// todo запросить пользователя отправить юзер-объект
		if err = appCtx.UserService.UpdatePartner(ctx, chatID, partnerName); err != nil {
			log.Print("handlers: failed to update partner: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Упс, не смог добавить твоего партнера 😿\n\n" +
					"Проверь правильность написания и попробуй ещё.\n\n" +
					"Если не поможет, то возможно это связано с тем, что у нас ещё не было диалога. " +
					"Буду ждать, пока вы оба зарегистрируетесь 🤗\n\n" +
					"Если и это не поможет, обратитесь к администратору",
			})
			return
		}

		sess.State = services.StateSettingsMenu
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf("Партнёр сохранён: %s", text),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
		})
	}
}
