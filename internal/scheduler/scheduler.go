package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/go-telegram/bot"
)

func StartPlanScheduler(
	ctx context.Context,
	b *bot.Bot,
	appCtx *app.AppContext,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for now := range ticker.C {
			duePlans, err := appCtx.PlanService.GetDueAndMark(now)
			if err != nil {
				log.Printf("scheduler: failed to fetch plans: %v", err)
				continue
			}
			for _, p := range duePlans {
				text := fmt.Sprintf(
					"📢Напоминание: %s (%s)",
					p.Description,
					appCtx.DateTimeService.FormatDateRu(p.EventTime.In(appCtx.Cfg.DefaultTZ)),
				)
				b.SendMessage(context.Background(), &bot.SendMessageParams{
					ChatID: p.ChatID,
					Text:   text,
				})
			}
		}
	}()
}
