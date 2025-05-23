package keyboards

import (
	"fmt"
	"strconv"

	"github.com/Petryanin/love-bot/internal/config"
	"github.com/Petryanin/love-bot/internal/db"
	"github.com/go-telegram/bot/models"
)

func PlansListInlineKeyboard(
	plans []db.Plan,
	pageNumber int,
	hasPrev, hasNext bool,
) *models.InlineKeyboardMarkup {
	buttons := make([]models.InlineKeyboardButton, len(plans))
	for i, p := range plans {
		idx := pageNumber*config.NavPageSize + i + 1
		buttons[i] = models.InlineKeyboardButton{
			Text:         strconv.Itoa(idx),
			CallbackData: fmt.Sprintf("plan:%d", p.ID),
		}
	}

	nav := make([]models.InlineKeyboardButton, 0, 2)
	if hasPrev {
		nav = append(nav, models.InlineKeyboardButton{
			Text:         config.BackArrowInlineBtn,
			CallbackData: fmt.Sprintf("plans:page:%d", pageNumber-1),
		})
	}
	if hasNext {
		nav = append(nav, models.InlineKeyboardButton{
			Text:         config.ForwardArrowInlineBtn,
			CallbackData: fmt.Sprintf("plans:page:%d", pageNumber+1),
		})
	}

	keyboard := [][]models.InlineKeyboardButton{buttons}
	if len(nav) > 0 {
		keyboard = append(keyboard, nav)
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func PlansDetailInlineKeyboard(plan *db.Plan, isRemindMenu bool) *models.InlineKeyboardMarkup {
	buttons := make([]models.InlineKeyboardButton, 0, 2)
	buttons = append(buttons, models.InlineKeyboardButton{
		Text:         config.DeleteInlineBtn,
		CallbackData: fmt.Sprintf("plan_delete:%d", plan.ID),
	})

	if isRemindMenu {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         config.BackInlineBtn,
			CallbackData: fmt.Sprintf("remind:%d", plan.ID),
		})
	} else {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         config.ToListInlineBtn,
			CallbackData: "plans",
		})
	}
	keyboard := [][]models.InlineKeyboardButton{buttons}
	return &models.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func PlansDeletedInlineKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: config.ToListInlineBtn, CallbackData: "plans"},
			},
		},
	}
}

func PlansReminderKeyboard(planID int64) *models.InlineKeyboardMarkup {
	deltas := []struct {
		Label string
		Min   int
	}{
		{"15м", 15},
		{"30м", 30},
		{"1ч", 60},
		{"3ч", 180},
		{"1д", 1440},
	}
	row1 := make([]models.InlineKeyboardButton, len(deltas))
	for i, d := range deltas {
		row1[i] = models.InlineKeyboardButton{
			Text:         d.Label,
			CallbackData: fmt.Sprintf("remind:change:%d:%d", planID, d.Min),
		}
	}

	row2 := []models.InlineKeyboardButton{
		{
			Text:         config.InputTimeInlineBtn,
			CallbackData: fmt.Sprintf("remind:change:%d:custom", planID),
		},
		{
			Text:         config.OpenInlineBtn,
			CallbackData: fmt.Sprintf("plan:%d:remind", planID),
		},
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			row1, row2,
		},
	}
}

func PlansOpenReminderKeyboard(planID int64) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{
				Text:         config.OpenInlineBtn,
				CallbackData: fmt.Sprintf("plan:%d:remind", planID),
			},
			{
				Text:         config.BackInlineBtn,
				CallbackData: fmt.Sprintf("remind:%d", planID),
			},
		}},
	}
}
