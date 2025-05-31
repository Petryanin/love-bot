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

func PlansHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)
		text := upd.Message.Text

		var allowedMap = map[string]bool{
			config.PlansBtn:  true,
			config.BackBtn:   true,
			config.AddBtn:    true,
			config.ListBtn:   true,
			config.CancelBtn: true,
		}

		if sess.State == services.StatePlanMenu && !allowedMap[text] {
			FallbackHandler(keyboards.PlanMenuKeyboard())(ctx, b, upd)
			return
		}

		switch sess.State {
		case services.StateRoot:
			sess.State = services.StatePlanMenu
			fallthrough

		case services.StatePlanMenu:
			plansMenuHandler(appCtx)(ctx, b, upd)
			return

		case services.StatePlanAddingAwaitDesc:
			plansAddingAwaitDescHandler(appCtx)(ctx, b, upd)
			return

		case services.StatePlanAddingAwaitEventTime:
			plansAddingAwaitEventTimeHandler(appCtx)(ctx, b, upd)
			return

		case services.StatePlanAddingAwaitRemindTime:
			plansAddingAwaitRemindTimeHandler(appCtx)(ctx, b, upd)
			return
		}

		FallbackHandler(keyboards.PlanMenuKeyboard())(ctx, b, upd)
	}
}

func plansMenuHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.PlansBtn || text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Меню планов: о чем вам напомнить?",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}
		// Добавить новый план
		if text == config.AddBtn {
			sess.State = services.StatePlanAddingAwaitDesc
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Введите текст напоминания",
				ReplyMarkup: keyboards.CancelKeyboard(),
			})
			return
		}
		// Список планов
		if text == config.ListBtn {
			PlansListHandler(appCtx)(ctx, b, upd)
		}
		if text == config.BackBtn {
			appCtx.SessionManager.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)
		}
	}
}

func plansAddingAwaitDescHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}
		sess.TempDesc = text
		sess.State = services.StatePlanAddingAwaitEventTime
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Когда это событие произойдёт?",
		})
	}
}

func plansAddingAwaitEventTimeHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}

		tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))

		parsedDT, err := appCtx.DateTimeService.ParseDateTime(ctx, text, time.Now(), tz.String())
		if err != nil {
			log.Print(err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё",
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

func plansAddingAwaitRemindTimeHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.CancelBtn {
			sess.State = services.StatePlanMenu
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}

		var remind time.Time
		if text == config.SameTimeBtn {
			remind = sess.TempEvent
		} else {
			tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))
			parsedDT, err := appCtx.DateTimeService.ParseDateTime(ctx, text, time.Now(), tz.String())
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё",
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
		if err := appCtx.PlanService.Add(p); err != nil {
			log.Print("handlers: failed to save plan: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "😥Ошибка при сохранении"})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "✅План сохранён!",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})

			partner, err := appCtx.UserService.Get(ctx, db.WithPartnerID(chatID))
			if err != nil {
				log.Print(err)
			} else {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: partner.ChatID,
					Text: fmt.Sprintf(
						"Твоя Вкущуща создала новый план: %s на %s",
						p.Description,
						appCtx.DateTimeService.FormatDateRu(p.EventTime.In(partner.TZ))),
				})
			}
		}
		sess.State = services.StatePlanMenu
	}
}

func PlansDetailsHandler(appCtx *app.AppContext) bot.HandlerFunc {
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

		plan, err := appCtx.PlanService.GetByID(planID, appCtx.Cfg)
		if err != nil {
			// todo add error handler
			log.Print("handlers: failed to get plan from DB: %w", err)
		}

		tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))

		replyText := strings.Join([]string{
			plan.Description + "\n",
			fmt.Sprintf("Время отправки уведомления:\n%s\n", plan.RemindTime.In(tz).Format(config.DTLayout)),
			fmt.Sprintf("Дата события:\n%s", plan.EventTime.In(tz).Format(config.DTLayout)),
		}, "\n")

		isRemindMenu := len(splitCallbackData) == 3

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			MessageID:   upd.CallbackQuery.Message.Message.ID,
			ChatID:      chatID,
			Text:        replyText,
			ReplyMarkup: keyboards.PlansDetailInlineKeyboard(plan, isRemindMenu),
		})
	}
}

func PlansDeleteHandler(appCtx *app.AppContext) bot.HandlerFunc {
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

		if err := appCtx.PlanService.Delete(planID); err != nil {
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

func PlansListHandler(appCtx *app.AppContext) bot.HandlerFunc {
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

		sess := appCtx.SessionManager.Get(chatID)
		sess.State = services.StatePlanMenu

		if upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, "plans:page:") {
			fmt.Sscanf(upd.CallbackQuery.Data, "plans:page:%d", &sess.TempPage)
		} else {
			sess.TempPage = 0
		}

		plans, hasPrev, hasNext, _ := appCtx.PlanService.List(sess.TempPage)
		if len(plans) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "У вас нет планов"})
			return
		}

		tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))

		var lines []string
		for i, p := range plans {
			lines = append(lines,
				fmt.Sprintf("%d) %s (%s)",
					sess.TempPage*config.NavPageSize+i+1,
					p.Description,
					appCtx.DateTimeService.FormatDateRu(p.EventTime.In(tz)),
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

func PlansChangeRemindTimeHandler(appCtx *app.AppContext) bot.HandlerFunc {
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

			sess := appCtx.SessionManager.Get(chatID)
			sess.State = services.StatePlanAddingAwaitRemindTime
			sess.TempPlanID = planID
			return
		}

		mins, _ := strconv.Atoi(arg)
		remindAt := time.Now().Add(time.Duration(mins) * time.Minute)

		plan, err := appCtx.PlanService.GetByID(planID, appCtx.Cfg)
		if err != nil {
			log.Print("handlers: failed to get plan from db: %w", err)
		}

		err = appCtx.PlanService.Schedule(plan.ID, remindAt)
		if err != nil {
			log.Print("handlers: failed to schedule plan: %w", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "😥Ошибка, не удалось сохранить новое время",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}

		tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))

		text := fmt.Sprintf(
			"%s\n\nХорошо, напомню вам снова в это время: %s",
			plan.Description,
			appCtx.DateTimeService.FormatDateRu(remindAt.In(tz)),
		)

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ReplyMarkup: keyboards.PlansOpenReminderKeyboard(planID),
		})
	}
}

func PlansRemindHandler(plan *db.Plan, appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		var chatID int64

		if upd != nil {
			b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: upd.CallbackQuery.ID,
				ShowAlert:       false,
			})

			parts := strings.Split(upd.CallbackQuery.Data, ":")
			planID, _ := strconv.ParseInt(parts[1], 10, 64)

			p, err := appCtx.PlanService.GetByID(planID, appCtx.Cfg)
			if err != nil {
				log.Print("handlers: failed to get plan from db: %w", err)
				return
			}
			plan = p
			chatID = upd.CallbackQuery.Message.Message.Chat.ID
		} else {
			chatID = plan.ChatID
		}

		tz := appCtx.UserService.TZ(ctx, appCtx.Cfg.DefaultTZ, db.WithChatID(chatID))

		text := fmt.Sprintf(
			"📢Напоминание: %s (%s)\n\n Напомнить снова через:",
			plan.Description,
			appCtx.DateTimeService.FormatDateRu(plan.EventTime.In(tz)),
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
			partner, err := appCtx.UserService.Get(ctx, db.WithPartnerID(chatID))
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
