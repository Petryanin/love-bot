package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/keyboards"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func PlansHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		sess := sm.Get(chatID)
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
			plansMenuHandler(cfg, sm, ps)(ctx, b, upd)
			return
		// Ввод описания
		case services.StateAddingAwaitDesc:
			plansAddingAwaitDescHandler(cfg, sm, ps)(ctx, b, upd)
			return
		// Ввод времени события
		case services.StateAddingAwaitEventTime:
			plansAddingAwaitEventTimeHandler(cfg, sm, ps)(ctx, b, upd)
			return
			// Ввод времени напоминания
		case services.StateAddingAwaitRemindTime:
			plansAddingAwaitRemindTimeHandler(cfg, sm, ps)(ctx, b, upd)
			return
		}

		// ничего больше не попало — fallback
		FallbackHandler(ctx, b, upd)
	}
}

func plansMenuHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := sm.Get(chatID)

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
			PlansListHandler(cfg, sm, ps)(ctx, b, upd)
		}
		if text == config.BackButton {
			sm.Reset(chatID)
			DefaultReplyHandler(ctx, b, upd)
		}
	}
}

func plansAddingAwaitDescHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := sm.Get(chatID)

		if text == config.CancelButton {
			sm.Reset(chatID)
			PlansHandler(cfg, sm, ps)(ctx, b, upd)
			return
		}
		sess.TempDesc = text
		sess.State = services.StateAddingAwaitEventTime
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Когда это событие произойдёт? (Укажи в формате DD-MM-YYYY HH:MM)",
		})
	}
}

func plansAddingAwaitEventTimeHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := sm.Get(chatID)

		if text == config.CancelButton {
			sm.Reset(chatID)
			PlansHandler(cfg, sm, ps)(ctx, b, upd)
			return
		}

		t, err := time.ParseInLocation(config.DTLayout, text, cfg.DefaultTZ)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё. Формат должен быть DD-MM-YYYY HH:MM",
			})
			return
		}
		sess.TempEvent = t
		sess.State = services.StateAddingAwaitRemindTime
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf(
                "Когда напомнить? (формат DD-MM-YYYY HH:MM, или можешь нажать %s)",
                config.SameTimeButton,
            ),
            ReplyMarkup: keyboards.PlanMenuRemindKeyboard(),
		})
	}
}

func plansAddingAwaitRemindTimeHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		text := upd.Message.Text
		sess := sm.Get(chatID)

		if text == config.CancelButton {
			sm.Reset(chatID)
			PlansHandler(cfg, sm, ps)(ctx, b, upd)
			return
		}

		var remind time.Time
		if text == config.SameTimeButton {
			remind = sess.TempEvent
		} else {
			r, err := time.ParseInLocation(config.DTLayout, text, cfg.DefaultTZ)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID, Text: "🧐Не смог распознать формат, попробуй ещё. Формат должен быть DD-MM-YYYY HH:MM",
				})
				return
			}
			remind = r
		}
		sess.TempRemind = remind

		// сохраняем в БД
		p := &services.Plan{
			ChatID:      chatID,
			Description: sess.TempDesc,
			EventTime:   sess.TempEvent,
			RemindTime:  sess.TempRemind,
		}
		if err := ps.Add(p); err != nil {
			log.Fatal(err)
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "😥Ошибка при сохранении"})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID, Text: "✅План сохранён!",
				ReplyMarkup: keyboards.PlanMenuKeyboard(),
			})
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: ps.PartnersChatIDs,
				Text:   fmt.Sprintf("Твоя любимка создала новый план: %s в %s", p.Description, p.EventTime.Format("02 Jan 2006 15:04")),
			})
		}
		sess.State = services.StateMenu
	}
}

func PlansDetailsHandler(cfg *config.Config, ps *services.PlanService) bot.HandlerFunc {
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

		plan, err := ps.GetByID(planID, cfg)
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

func PlansDeleteHandler(cfg *config.Config, ps *services.PlanService) bot.HandlerFunc {
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

		if err := ps.Delete(planID); err != nil {
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

func PlansListHandler(cfg *config.Config, sm *services.SessionManager, ps *services.PlanService) bot.HandlerFunc {
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

		sess := sm.Get(chatID)
		sess.State = services.StateMenu
		plans, _ := ps.List(chatID, cfg)
		if len(plans) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "У вас нет планов"})
			sess.State = services.StateMenu
			return
		}

		var lines []string
		for i, p := range plans {
			lines = append(lines,
				fmt.Sprintf("%d) %s (%s)",
					i + 1,
					p.Description,
					p.EventTime.Format(config.DTLayout), // или любой ваш формат
				),
			)
		}
		msgText := strings.Join(lines, "\n") + "\n\nВыбери план для подробностей:"
        kb := keyboards.PlansListInlineKeyboard(plans)

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
