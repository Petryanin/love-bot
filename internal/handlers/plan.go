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
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func PlansHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		sess := appCtx.SessionManager.Get(chatID)
		text := upd.Message.Text

		var allowedMap = map[string]struct{}{
			config.PlansButton:  {},
			config.BackButton:   {},
			config.AddButton:    {},
			config.ListButton:   {},
			config.CancelButton: {},
		}
		_, ok := allowedMap[text]
		if sess.State == services.StateMenu && !ok {
			FallbackHandler(ctx, b, upd)
			return
		}

		switch sess.State {
		// Главное меню “Планы”
		case services.StateMenu:
			plansMenuHandler(appCtx)(ctx, b, upd)
			return
		// Ввод описания
		case services.StateAddingAwaitDesc:
			plansAddingAwaitDescHandler(appCtx)(ctx, b, upd)
			return
		// Ввод времени события
		case services.StateAddingAwaitEventTime:
			plansAddingAwaitEventTimeHandler(appCtx)(ctx, b, upd)
			return
			// Ввод времени напоминания
		case services.StateAddingAwaitRemindTime:
			plansAddingAwaitRemindTimeHandler(appCtx)(ctx, b, upd)
			return
		}

		// ничего больше не попало — fallback
		FallbackHandler(ctx, b, upd)
	}
}

func plansMenuHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := appCtx.SessionManager.Get(chatID)

		if text == config.PlansButton || text == config.CancelButton {
			sess.State = services.StateMenu
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Меню планов: о чем вам напомнить?",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			return
		}
		// Добавить новый план
		if text == config.AddButton {
			sess.State = services.StateAddingAwaitDesc
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "Введите текст напоминания",
				ReplyMarkup: keyboards.PlanMenuCancelKeyboard(),
			})
			return
		}
		// Список планов
		if text == config.ListButton {
			PlansListHandler(appCtx)(ctx, b, upd)
		}
		if text == config.BackButton {
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

		if text == config.CancelButton {
			appCtx.SessionManager.Reset(chatID)
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}
		sess.TempDesc = text
		sess.State = services.StateAddingAwaitEventTime
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

		if text == config.CancelButton {
			appCtx.SessionManager.Reset(chatID)
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}

		parsedDT, err := appCtx.DateTimeService.ParseDateTime(text, time.Now())
		if err != nil {
			log.Print(err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё",
			})
			return
		}

		sess.TempEvent = parsedDT
		sess.State = services.StateAddingAwaitRemindTime

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

		if text == config.CancelButton {
			appCtx.SessionManager.Reset(chatID)
			PlansHandler(appCtx)(ctx, b, upd)
			return
		}

		var remind time.Time
		if text == config.SameTimeButton {
			remind = sess.TempEvent
		} else {
			parsedDT, err := appCtx.DateTimeService.ParseDateTime(text, time.Now())
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
		p := &services.Plan{
			ChatID:      chatID,
			Description: sess.TempDesc,
			EventTime:   sess.TempEvent,
			RemindTime:  sess.TempRemind,
		}
		if err := appCtx.PlanService.Add(p); err != nil {
			log.Fatal(err)
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "😥Ошибка при сохранении"})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        "✅План сохранён!",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})

			for _, id := range appCtx.PlanService.PartnersChatIDs {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: id,
					Text: fmt.Sprintf(
						"Твоя Вкущуща создала новый план: %s на %s",
						p.Description,
						appCtx.DateTimeService.FormatDateRu(p.EventTime)),
				})
			}
		}
		sess.State = services.StateMenu
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

		planID, err := strconv.ParseInt(strings.Split(callbackData, ":")[1], 10, 64)
		if err != nil {
			// todo add error handler
			log.Print("failed to get planID from callback data: %w", err)
		}

		plan, err := appCtx.PlanService.GetByID(planID, appCtx.Cfg)
		if err != nil {
			// todo add error handler
			log.Print("failed to get plan from DB: %w", err)
		}

		replyText := strings.Join([]string{
			plan.Description + "\n",
			fmt.Sprintf("Время отправки уведомления:\n%s\n", plan.RemindTime.Format(config.DTLayout)),
			fmt.Sprintf("Дата события:\n%s", plan.EventTime.Format(config.DTLayout)),
		}, "\n")

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			MessageID:   upd.CallbackQuery.Message.Message.ID,
			ChatID:      chatID,
			Text:        replyText,
			ReplyMarkup: keyboards.PlansDetailInlineKeyboard(plan),
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
		sess.State = services.StateMenu

		if upd.CallbackQuery != nil && strings.HasPrefix(upd.CallbackQuery.Data, "plans:page:") {
			fmt.Sscanf(upd.CallbackQuery.Data, "plans:page:%d", &sess.TempPage)
		} else {
			sess.TempPage = 0
		}

		plans, hasPrev, hasNext, _ := appCtx.PlanService.List(chatID, sess.TempPage, config.NavPageSize, appCtx.Cfg)
		if len(plans) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "У вас нет планов"})
			return
		}

		var lines []string
		for i, p := range plans {
			lines = append(lines,
				fmt.Sprintf("%d) %s (%s)",
					i+1,
					p.Description,
					appCtx.DateTimeService.FormatDateRu(p.EventTime),
				),
			)
		}

		msgText := strings.Join(lines, "\n") + "\n\nВыбери план для подробностей:"
		kb := keyboards.PlansListInlineKeyboard(plans, sess.TempPage, config.NavPageSize, hasPrev, hasNext)

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
