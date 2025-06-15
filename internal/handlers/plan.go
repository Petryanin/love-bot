package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

func PlansHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := app.Session.Get(chatID)

		switch sess.State {
		case services.StateRoot:
			sess.State = services.StatePlanMenu
			fallthrough

		case services.StatePlanMenu:
			plansMenuHandler(app)(ctx, b, upd)

		case services.StatePlanAddingAwaitDesc:
			plansAddingAwaitDescHandler(app)(ctx, b, upd)

		case services.StatePlanAddingAwaitEventTime:
			plansAddingAwaitEventTimeHandler(app)(ctx, b, upd)

		case services.StatePlanAddingAwaitRemindTime:
			plansAddingAwaitRemindTimeHandler(app)(ctx, b, upd)

		default:
			FallbackHandler(keyboards.PlanMenuKeyboard())(ctx, b, upd)

		}
	}
}

func plansMenuHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		switch {
		case text == config.PlansBtn || text == config.CancelBtn:
			sess.State = services.StatePlanMenu
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Меню планов: о чем вам напомнить?",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})

		case text == config.AddBtn:
			sess.State = services.StatePlanAddingAwaitDesc
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Введите текст напоминания",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

		case text == config.ListBtn:
			PlansListHandler(app)(ctx, b, upd)

		case text == config.BackBtn:
			app.Session.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)

		default:
			plansAddingAwaitDescHandler(app)(ctx, b, upd)
		}
	}
}

func plansAddingAwaitDescHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(app)(ctx, b, upd)
			return
		}

		// пробуем сразу распознать время
		now := time.Now()
		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))
		parsedBody, parsedDT, err := app.DateTime.Parse(ctx, text, now, tz.String())
		if err == nil {
			desc := strings.TrimSpace(strings.Replace(text, parsedBody, "", 1))
			if desc != "" {
				if parsedDT.Before(now) {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: chatID,
						Text:   "🧐Время должно быть в будущем, попробуй ещё",
					})
					return
				}

				sess.TempDesc = desc
				sess.TempEvent = parsedDT
				sess.State = services.StatePlanAddingAwaitRemindTime

				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:      chatID,
					Text:        "Когда напомнить?",
					ReplyMarkup: keyboards.PlanMenuRemindKeyboard(),
				})
				return
			}
		}

		sess.TempDesc = text
		sess.State = services.StatePlanAddingAwaitEventTime
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Когда это событие произойдёт?",
			ReplyMarkup: keyboards.CancelKeyboard(),
		})
	}
}

func plansAddingAwaitEventTimeHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(app)(ctx, b, upd)
			return
		}

		now := time.Now()
		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))

		_, parsedDT, err := app.DateTime.Parse(ctx, text, now, tz.String())
		if err != nil {
			log.Print(err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё",
			})
			return
		}

		if parsedDT.Before(now) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Время должно быть в будущем, попробуй ещё",
			})
			return
		}

		sess.TempEvent = parsedDT
		sess.State = services.StatePlanAddingAwaitRemindTime

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Когда напомнить?",
			ReplyMarkup: keyboards.PlanMenuRemindKeyboard(),
		})
	}
}

func plansAddingAwaitRemindTimeHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := app.Session.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(app)(ctx, b, upd)
			return
		}

		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))
		var remind time.Time

		if text == config.SameTimeBtn {
			remind = sess.TempEvent
		} else {
			now := time.Now()

			_, parsedDT, err := app.DateTime.Parse(ctx, text, now, tz.String())
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё",
				})
				return
			}
			if parsedDT.Before(now) {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Время должно быть в будущем, попробуй ещё",
				})
				return
			}
			if parsedDT.After(sess.TempEvent) {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Время напоминания не может быть больше времени события, попробуй ещё",
				})
				return
			}
			remind = parsedDT
		}
		sess.TempRemind = remind

		// сохраняем в БД
		p := &db.Plan{
			ChatID:      chatID,
			Description: sess.TempDesc,
			EventTime:   sess.TempEvent,
			RemindTime:  sess.TempRemind,
		}
		if err := app.Plan.Add(ctx, p); err != nil {
			log.Print("handlers: failed to save plan: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "😥Ошибка при сохранении"})
		} else {
			msg := fmt.Sprintf(
				"✅План сохранён!\n\n%s %s\n(напомню %s)",
				p.Description,
				app.DateTime.FormatRu(p.EventTime.In(tz)),
				app.DateTime.FormatRu(p.RemindTime.In(tz)),
			)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        msg,
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})

			partner, err := app.User.Get(ctx, db.WithPartnerID(chatID))
			if err != nil {
				log.Print(err)
			} else {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: partner.ChatID,
					Text: fmt.Sprintf(
						"Твоя Вкущуща создала новый план: %s на %s",
						p.Description,
						app.DateTime.FormatRu(p.EventTime.In(partner.TZ))),
				})
			}
		}
		sess.State = services.StatePlanMenu
	}
}

func PlansDetailsHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: upd.CallbackQuery.ID,
			ShowAlert:       false,
		})

		chatID := upd.CallbackQuery.Message.Message.Chat.ID
		callbackData := upd.CallbackQuery.Data

		splitCallbackData := strings.Split(callbackData, ":")
		planID, err := strconv.ParseInt(splitCallbackData[1], 10, 64)
		if err != nil {
			// todo add error handler
			log.Print("handlers: failed to get planID from callback data: %w", err)
		}

		plan, err := app.Plan.GetByID(ctx, planID, app.Cfg)
		if err != nil {
			// todo add error handler
			log.Print("handlers: failed to get plan from DB: %w", err)
		}

		user, err := app.User.Get(ctx, db.WithChatID(chatID))
		if err != nil {
			log.Print(err)
			return
		}

		var author *db.UserFull
		var canDelete bool
		if plan.ChatID == chatID {
			author = user
			canDelete = true
		} else {
			author, err = app.User.Get(ctx, db.WithChatID(plan.ChatID))
			canDelete = false
		}

		replyText := strings.Join([]string{
			plan.Description + "\n",
			fmt.Sprintf("Время отправки уведомления:\n%s\n", plan.RemindTime.In(user.TZ).Format(config.DTLayout)),
			fmt.Sprintf("Дата события:\n%s\n", plan.EventTime.In(user.TZ).Format(config.DTLayout)),
			fmt.Sprintf("Создан: @%s", author.Name),
		}, "\n")

		isRemindMenu := len(splitCallbackData) == 3

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			MessageID:   upd.CallbackQuery.Message.Message.ID,
			ChatID:      chatID,
			Text:        replyText,
			ReplyMarkup: keyboards.PlansDetailInlineKeyboard(plan, canDelete, isRemindMenu),
		})
	}
}

func PlansDeleteHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: upd.CallbackQuery.ID,
			ShowAlert:       false,
		})

		chatID := upd.CallbackQuery.Message.Message.Chat.ID
		callbackData := upd.CallbackQuery.Data

		planID, err := strconv.ParseInt(strings.Split(callbackData, ":")[1], 10, 64)
		if err != nil {
			// todo add error handler
			log.Print("failed to get planID from callback data: %w", err)
		}

		if err := app.Plan.Delete(ctx, planID); err != nil {
			// todo add error handler
			log.Print("failed to get plan from DB: %w", err)
		}

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			MessageID:   upd.CallbackQuery.Message.Message.ID,
			ChatID:      chatID,
			Text:        "👌Напоминание успешно удалено",
			ReplyMarkup: keyboards.PlansDeletedInlineKeyboard(),
		})
	}
}

func PlansListHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		var callbackMsgID int
		var chatID int64

		if upd.Message != nil {
			chatID = upd.Message.Chat.ID
		} else {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: upd.CallbackQuery.ID,
				ShowAlert:       false,
			})
			chatID = upd.CallbackQuery.Message.Message.Chat.ID
			callbackMsgID = upd.CallbackQuery.Message.Message.ID
		}

		sess := app.Session.Get(chatID)
		sess.State = services.StatePlanMenu

		if upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, "plans:page:") {
			fmt.Sscanf(upd.CallbackQuery.Data, "plans:page:%d", &sess.TempPage)
		} else {
			sess.TempPage = 0
		}

		plans, hasPrev, hasNext, _ := app.Plan.List(ctx, sess.TempPage)
		if len(plans) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "У вас нет планов"})
			return
		}

		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))

		var lines []string
		for i, p := range plans {
			lines = append(lines,
				fmt.Sprintf("%d) %s (%s)",
					sess.TempPage*config.NavPageSize+i+1,
					p.Description,
					app.DateTime.FormatRu(p.EventTime.In(tz)),
				),
			)
		}

		msgText := strings.Join(lines, "\n") + "\n\nВыбери план для подробностей:"
		kb := keyboards.PlansListInlineKeyboard(plans, sess.TempPage, hasPrev, hasNext)

		if callbackMsgID != 0 {
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				MessageID:   callbackMsgID,
				ChatID:      chatID,
				Text:        msgText,
				ReplyMarkup: kb,
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        msgText,
				ReplyMarkup: kb,
			})
		}
	}
}

func PlansChangeRemindTimeHandler(app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		cb := upd.CallbackQuery
		data := cb.Data
		parts := strings.Split(data, ":")

		planID, _ := strconv.ParseInt(parts[2], 10, 64)
		chatID := cb.Message.Message.Chat.ID
		messageID := cb.Message.Message.ID

		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cb.ID,
			ShowAlert:       false,
		})

		arg := parts[3]
		if arg == "custom" {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Введите время повторного напоминания",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})

			sess := app.Session.Get(chatID)
			sess.State = services.StatePlanAddingAwaitRemindTime
			sess.TempPlanID = planID
			return
		}

		plan, err := app.Plan.GetByID(ctx, planID, app.Cfg)
		if err != nil {
			log.Print("handlers: failed to get plan from db: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "🧐Такого плана не найдено",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}

		mins, _ := strconv.Atoi(arg)
		remindAt := time.Now().Add(time.Duration(mins) * time.Minute)
		if remindAt.After(plan.EventTime) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "🧐Время напоминания не может быть больше времени события, попробуй ещё",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}

		err = app.Plan.Schedule(ctx, plan.ID, remindAt)
		if err != nil {
			log.Print("handlers: failed to schedule plan: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "😥Ошибка, не удалось сохранить новое время",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}

		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))

		text := fmt.Sprintf(
			"%s\n\nХорошо, напомню вам снова в это время: %s",
			plan.Description,
			app.DateTime.FormatRu(remindAt.In(tz)),
		)

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ReplyMarkup: keyboards.PlansOpenReminderKeyboard(planID),
		})
	}
}

func PlansRemindHandler(plan *db.Plan, app *app.App) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		var chatID int64

		if upd != nil {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: upd.CallbackQuery.ID,
				ShowAlert:       false,
			})

			parts := strings.Split(upd.CallbackQuery.Data, ":")
			planID, _ := strconv.ParseInt(parts[1], 10, 64)

			p, err := app.Plan.GetByID(ctx, planID, app.Cfg)
			if err != nil {
				log.Print("handlers: failed to get plan from db: %w", err)
				return
			}
			plan = p
			chatID = upd.CallbackQuery.Message.Message.Chat.ID
		} else {
			chatID = plan.ChatID
		}

		tz := app.User.TZ(ctx, app.Cfg.DefaultTZ, db.WithChatID(chatID))

		text := fmt.Sprintf(
			"📢Напоминание: %s (%s)\n\n Напомнить снова через:",
			plan.Description,
			app.DateTime.FormatRu(plan.EventTime.In(tz)),
		)

		kb := keyboards.PlansReminderKeyboard(plan.ID)
		if upd != nil {
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   upd.CallbackQuery.Message.Message.ID,
				Text:        text,
				ReplyMarkup: kb,
			})
		} else {
			chatIDs := []int64{chatID}
			partner, err := app.User.Get(ctx, db.WithPartnerID(chatID))
			if err != nil {
				log.Print(err)
			} else {
				chatIDs = append(chatIDs, partner.ChatID)
			}

			for _, id := range chatIDs {
				b.SendMessage(context.Background(), &bot.SendMessageParams{
					ChatID:      id,
					Text:        text,
					ReplyMarkup: kb,
				})
			}
		}
	}
}
