package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
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

func SettingsHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)
		text := upd.Message.Text

		var allowedMap = map[string]bool{
			config.CityBtn:    true,
			config.PartnerBtn: true,
			config.CatBtn:     true,
			config.DisableBtn: true,
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
			settingsMenuHandler(app)(ctx, b, upd)
			return

		case services.StateSettingsCity:
			settingsCityHandler(app)(ctx, b, upd)
			return

		case services.StateSettingsPartner:
			settingsPartnerHandler(app)(ctx, b, upd)
			return

		case services.StateSettingsCat:
			settingsCatHandler(app)(ctx, b, upd)
			return
		}

		FallbackHandler(keyboards.SettingsMenuKeyboard())(ctx, b, upd)
	}
}

func settingsMenuHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		switch {
		case text == config.SettingsBtn || text == config.CancelBtn:
			user, err := app.User.Get(ctx, db.WithChatID(chatID), db.WithPartnerInfo())
			if err != nil {
				log.Print("handlers: failed to get user info: %w", err)
				app.Session.Reset(chatID)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:      chatID,
					Text:        "Упс, не удалось получить информацию по пользователю 😿\nПопробуй позже",
					ReplyMarkup: keyboards.BaseReplyKeyboard(),
				})
				return
			}

			tz := user.TZ.String()

			var catTimeStr string
			catTime := user.CatTime
			if catTime.IsZero() {
				catTimeStr = "Отключены"
			} else {
				catTimeStr = "Каждый день в " + catTime.Format("15:04")
			}
			msg := fmt.Sprintf(
				"*Ваши текущие настройки:*\n\n"+
					"\\- город: *%s*\n"+
					"\\- часовой пояс: *%s*\n"+
					"\\- партнер: @%s\n"+
					"\\- котики: *%s*",
				user.City, tz, bot.EscapeMarkdown(user.PartnerName), catTimeStr,
			)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        msg,
				ReplyMarkup: keyboards.SettingsMenuKeyboard(),
				ParseMode:   models.ParseModeMarkdown,
			})

		case text == config.CityBtn:
			sess.State = services.StateSettingsCity
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Пожалуйста, отправь мне название своего города или твою геолокацию.\n\n" +
					"Это поможет мне предоставлять актуальную сводку погоды и учитывать часовой пояс при напоминаниях",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

		case text == config.PartnerBtn:
			sess.State = services.StateSettingsPartner
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Пожалуйста, введи Telegram-ник твоего партнёра.\n\n" +
					"Это поможет мне учитывать ваши совместные планы.",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

		case text == config.CatBtn:
			sess.State = services.StateSettingsCat
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Пожалуйста, введи время, в которое тебе ежедневно будут прилетать котики (в формате HH:MM) " +
					"или нажми «" + config.DisableBtn + "», чтобы отказаться от подписки.",
				ReplyMarkup: keyboards.DisableKeyboard(),
			})

		case text == config.BackBtn:
			app.Session.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)
		}
	}
}

func settingsCityHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		var city string
		var tz string
		var err error

		if text != "" {
			if text == config.CancelBtn {
				sess.State = services.StateSettingsMenu
				SettingsHandler(app)(ctx, b, upd)
				return
			}

			city, tz, err = app.Geo.ResolveByName(ctx, text)
			if err != nil {
				log.Print("handlers: failed to resolve geo info by city name: %w", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать город, попробуй ещё",
				})
				return
			}
		} else if upd.Message.Location != nil {
			city, tz, err = app.Geo.ResolveByCoords(ctx, upd.Message.Location.Latitude, upd.Message.Location.Longitude)
			if err != nil {
				log.Print("handlers: failed to resolve geo info by coords: %w", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать геолокацию, попробуй ещё",
				})
				return
			}
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Не могу распознять такой тип сообщения, попробуй ещё",
			})
			return
		}

		app.User.UpdateGeo(ctx, chatID, city, tz)
		sess.State = services.StateSettingsMenu

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf("Город сохранён: *%s*\nЧасовой пояс: *%s*", city, tz),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func settingsPartnerHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StateSettingsMenu
			SettingsHandler(app)(ctx, b, upd)
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
		if err = app.User.UpdatePartner(ctx, chatID, partnerName); err != nil {
			log.Print("handlers: failed to update partner: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "Упс, не смог добавить твоего партнера 😿\n\n" +
					"Проверь правильность написания и попробуй ещё.\n\n" +
					"Если не поможет, то возможно это связано с тем, что у нас ещё не было диалога. " +
					"Буду ждать, пока вы оба зарегистрируетесь 🤗\n\n" +
					"Если и это не поможет, обратись к администратору",
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

func settingsCatHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StateSettingsMenu
			SettingsHandler(app)(ctx, b, upd)
			return
		}

		if text == config.DisableBtn {
			err := app.User.UpdateCatTime(ctx, chatID, "")
			if err != nil {
				log.Print("handlers: %w", err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Упс, произошла ошибка. Обратись к администратору",
				})
				return
			}
			sess.State = services.StateSettingsMenu
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "👌Хорошо, больше не буду присылать тебе котиков",
				ReplyMarkup: keyboards.SettingsMenuKeyboard(),
			})
			return
		}

		user, err := app.User.Get(ctx, db.WithChatID(chatID))
		if err != nil {
			log.Print("handlers: failed to get user info: %w", err)
			app.Session.Reset(chatID)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Упс, не удалось получить информацию по пользователю 😿\nПопробуй позже",
				ReplyMarkup: keyboards.BaseReplyKeyboard(),
			})
			return
		}
		_, err = time.ParseInLocation("15:04", text, user.TZ)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "😿Неверный формат. Введи время в виде HH:MM, например «18:30»",
			})
			return
		}

		if err := app.User.UpdateCatTime(ctx, chatID, text); err != nil {
			log.Printf("handlers: UpdateCatTime error: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Упс, не удалось сохранить время 😿\nПопробуй позже",
			})
			return
		}
		sess.State = services.StateSettingsMenu
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        fmt.Sprintf("✅ Время сохранено! Ежедневно в %s тебе будут прилетать котики😻", text),
			ReplyMarkup: keyboards.SettingsMenuKeyboard(),
		})
		return
	}
}
