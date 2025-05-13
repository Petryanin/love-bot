package keyboards

import (
	"fmt"
	"strconv"

	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot/models"
)

func PlansListInlineKeyboard(plans []services.Plan) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, len(plans))
	for i, p := range plans {
		rows[i] = []models.InlineKeyboardButton{
			{
				Text:         strconv.Itoa(i + 1),
				CallbackData: fmt.Sprintf("plan:%d", p.ID),
			},
		}
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func PlansDetailInlineKeyboard(plan *services.Plan) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: config.DeleteButton, CallbackData: fmt.Sprintf("plan_delete:%d", plan.ID)},
			},
			{
				{Text: config.ReturnToListButton, CallbackData: "plan_list"},
			},
		},
	}
}

func PlansDeletedInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: config.ReturnToListButton, CallbackData: "plan_list"},
			},
		},
	}
}
