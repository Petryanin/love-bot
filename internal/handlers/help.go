package handlers

import (
	"context"
	"strings"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func HelpHandler(appCtx *app.AppContext) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		chatID := upd.Message.Chat.ID
		sess := appCtx.SessionManager.Get(chatID)

		if sess.State != services.StateRoot {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "В данный момент команда недоступна😢",
			})
			return
		}
		appCtx.SessionManager.Reset(chatID)

		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: chatID,
			Action: models.ChatActionTyping,
		})

		helpText := strings.Join([]string{
			"Вот что я умею:",
			"",
			"*Команды:*",
			"/start — приветствие и главное меню",
			"/help — показать это сообщение",
			"",
			"*Кнопки:*",
			"\"" + config.WeatherBtn + "\" — краткая сводка текущей погоды",
			"\"" + strings.Replace(config.ComplimentBtn, "-", "\\-", -1) + "\" — картинка с котом и комплиментом",
			"\"" + config.PlansBtn + "\" — меню планов и напоминаний",
			"\"" + config.TogetherTimeBtn + "\" — время ваших отношений",
			"\"" + config.MagicBallBtn + "\" — поможет тебе принять решение",
			"\"" + config.SettingsBtn + "\" — меню твоих настроек",
		}, "\n")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      helpText,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
