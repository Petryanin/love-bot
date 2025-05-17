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
				{Text: config.WeatherBtn},
				{Text: config.ComplimentBtn},
				{Text: config.PlansBtn},
			},
			{
				{Text: config.TogetherTimeBtn},
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
				{Text: config.AddBtn},
				{Text: config.ListBtn},
			},
			{
				{Text: config.BackBtn},
			},
		},
		ResizeKeyboard: true,
	}
}

func PlanMenuCancelKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.CancelBtn},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}

func PlanMenuRemindKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.SameTimeBtn},
			},
			{
				{Text: config.CancelBtn},
			},
		},
		ResizeKeyboard: true,
	}
}
