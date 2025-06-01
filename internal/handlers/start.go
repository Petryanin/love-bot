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
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "В данный момент команда недоступна😢",
			})
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

		case services.StateStartCity:
			startCityHandler(app)(ctx, b, upd)
			return

		case services.StateStartPartner:
			startPartnerHandler(app)(ctx, b, upd)
			return
		}

		text := "Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n" +
			bot.EscapeMarkdown(
				"Можешь сразу нажать на одну из кнопок или позвать команду /help, "+
					"чтобы ознакомиться с моими функциями подробнее.\n\n\n"+
					"От разработчика:\n"+
					"Меня зовут Алексей @Petryanin\n") +
			"Исходный код открыт и выложен на [github](https://github.com/Petryanin/love-bot)" +
			"Я очень старался и буду благодарен за ⭐"

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ReplyMarkup: keyboards.BaseReplyKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func StartSettingsHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)
		sess.State = services.StateStartCity

		text := "Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n" +
			bot.EscapeMarkdown("Для начала мне нужно узнать тебя получше. "+
				"Пожалуйста, отправь мне название своего города или твою геолокацию.\n\n"+
				"Это поможет мне предоставлять актуальную сводку погоды и учитывать часовой пояс при напоминаниях")
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
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		var city string
		var tz string
		var err error

		if text != "" {
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

		sess.TempStartCity = city
		sess.TempStartTZ = tz
		sess.State = services.StateStartPartner

		msg := fmt.Sprintf("Город сохранён: *%s*\nЧасовой пояс: *%s*", city, tz) +
			bot.EscapeMarkdown("\n\nТеперь отправь мне Telegram-ник твоего партнёра.\n\n"+
				"Это поможет мне учитывать ваши совместные планы.")
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
			log.Print("handlers: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text: "🧐Не смог распознать имя пользователя." +
					"Проверь правильность написания и попробуй ещё",
			})
			return
		}

		// todo запросить пользователя отправить юзер-объект
		// todo переделать логику регистрации партнеров
		partner, err := app.User.Get(ctx, db.WithUsername(partnerName))
		if err != nil {
			log.Print("handlers: failed to get partner: %w", err)
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

		// todo разделить логику
		err = app.User.Upsert(ctx, chatID, upd.Message.From.Username, sess.TempStartCity, sess.TempStartTZ, partner.ChatID)
		if err != nil {
			log.Print("handlers: failed to upsert partner: %w", err)
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
		app.Session.Reset(chatID)

		if !strings.HasPrefix(text, "@") {
			text = "@" + text
		}

		msg := bot.EscapeMarkdown(
			fmt.Sprintf("Партнёр сохранён: %s", text)+
				"\n\nСупер, всё готово!👍\n\n"+
				"Можешь сразу нажать на одну из кнопок или позвать команду /help, "+
				"чтобы ознакомиться с моими функциями подробнее.\n\n\n"+
				"От разработчика:\n"+
				"Меня зовут Алексей @Petryanin\n") +
			"Исходный код открыт и выложен на [github](https://github.com/Petryanin/love-bot)" +
			"Я очень старался и буду благодарен за ⭐"

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: keyboards.BaseReplyKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}
