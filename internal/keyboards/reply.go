// пока не используется

package keyboards

import (
	"github.com/Petryanin/love-bot/internal/config"
	"github.com/go-telegram/bot/models"
)

func BaseReplyKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.WeatherButton},
				{Text: config.ComplimentButton},
				{Text: config.PlansButton},
			},
			{
				{Text: config.TogetherTimeButton},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}
}

func PlanMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.AddButton},
				{Text: config.ListButton},
			},
			{
				{Text: config.BackButton},
			},
		},
		ResizeKeyboard: true,
	}
}

func PlanMenuCancelKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.CancelButton},
			},
		},
		ResizeKeyboard: true,
	}
}

func PlanMenuRemindKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.SameTimeButton},
			},
			{
				{Text: config.CancelButton},
			},
		},
		ResizeKeyboard: true,
	}
}
