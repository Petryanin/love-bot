package keyboards

import (
	"fmt"
	"strconv"

	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/services"
	"github.com/go-telegram/bot/models"
)

func PlansListInlineKeyboard(
	plans []services.Plan,
	pageNumber, pageSize int,
	hasPrev, hasNext bool,
) *models.InlineKeyboardMarkup {
	buttons := make([]models.InlineKeyboardButton, len(plans))
	for i, p := range plans {
		idx := pageNumber*pageSize + i + 1
		buttons[i] = models.InlineKeyboardButton{
			Text:         strconv.Itoa(idx),
			CallbackData: fmt.Sprintf("plan:%d", p.ID),
		}
	}

	nav := make([]models.InlineKeyboardButton, 0, 2)
	if hasPrev {
		nav = append(nav, models.InlineKeyboardButton{
			Text:         "<< Назад",
			CallbackData: fmt.Sprintf("plans:page:%d", pageNumber-1),
		})
	}
	if hasNext {
		nav = append(nav, models.InlineKeyboardButton{
			Text:         "Вперед >>",
			CallbackData: fmt.Sprintf("plans:page:%d", pageNumber+1),
		})
	}

	keyboard := [][]models.InlineKeyboardButton{buttons}
	if len(nav) > 0 {
		keyboard = append(keyboard, nav)
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func PlansDetailInlineKeyboard(plan *services.Plan) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: config.DeleteButton, CallbackData: fmt.Sprintf("plan_delete:%d", plan.ID)},
			},
			{
				{Text: config.ReturnToListButton, CallbackData: "plans"},
			},
		},
	}
}

func PlansDeletedInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: config.ReturnToListButton, CallbackData: "plans"},
			},
		},
	}
}
