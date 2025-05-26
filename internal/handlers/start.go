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

func StartHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)

		if sess.State != services.StateRoot && !appCtx.SessionManager.IsStartSettingsState(chatID) {
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
			appCtx.SessionManager.Reset(chatID)
			_, err := appCtx.UserService.Get(ctx, db.WithChatID(chatID))
			if err != nil {
				StartSettingsHandler(appCtx)(ctx, b, upd)
				return
			}

		case services.StateStartCity:
			startCityHandler(appCtx)(ctx, b, upd)
			return

		case services.StateStartPartner:
			startPartnerHandler(appCtx)(ctx, b, upd)
			return
		}

		codeLines := services.CountCodeLines(".")
		linesCount := fmt.Sprintf("%d %s", codeLines, services.Pluralize(codeLines, "строку", "строки", "строк"))

		text := "Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n" +
			bot.EscapeMarkdown(
				"Можешь сразу нажать на одну из кнопок или позвать команду /help, "+
					"чтобы ознакомиться с моими функциями подробнее.\n\n\n"+
					"От разработчика:\n"+
					"Меня зовут Алексей @Petryanin\n") +
			fmt.Sprintf("Исходный код открыт — на данный момент проект насчитывает *%s*\\.\n", linesCount) +
			"Я очень старался и буду благодарен за ⭐ на [github](https://github.com/Petryanin/love-bot)"

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ReplyMarkup: keyboards.BaseReplyKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}

func StartSettingsHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)
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

func startCityHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		var city string
		var tz string
		var err error

		if text != "" {
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

func startPartnerHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

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
		partner, err := appCtx.UserService.Get(ctx, db.WithUsername(partnerName))
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
		err = appCtx.UserService.Upsert(ctx, chatID, upd.Message.From.Username, sess.TempStartCity, sess.TempStartTZ, partner.ChatID)
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
		appCtx.SessionManager.Reset(chatID)

		codeLines := services.CountCodeLines(".")
		linesCount := fmt.Sprintf("%d %s", codeLines, services.Pluralize(codeLines, "строку", "строки", "строк"))

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
			fmt.Sprintf("Исходный код открыт — на данный момент проект насчитывает *%s*\\.\n", linesCount) +
			"Я очень старался и буду благодарен за ⭐ на [github](https://github.com/Petryanin/love-bot)"

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: keyboards.BaseReplyKeyboard(),
			ParseMode:   models.ParseModeMarkdown,
		})
	}
}
