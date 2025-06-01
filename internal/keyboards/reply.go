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
				{Text: config.MagicBallBtn},
				{Text: config.SettingsBtn},
			},
		},
		ResizeKeyboard: true,
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

func CancelKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.CancelBtn},
			},
		},
		ResizeKeyboard: true,
	}
}

func DisableKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.DisableBtn},
			},
			{
				{Text: config.CancelBtn},
			},
		},
		ResizeKeyboard: true,
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

func SettingsMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: config.CityBtn}, {Text: config.PartnerBtn}, {Text: config.CatBtn},
			},
			{
				{Text: config.BackBtn},
			},
		},
		ResizeKeyboard: true,
	}
}
