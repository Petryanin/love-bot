package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/Petryanin/love-bot/internal/app"
	"github.com/Petryanin/love-bot/internal/handlers"
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
				handlers.PlansRemindHandler(&p, appCtx)(ctx, b, nil)
			}
		}
	}()
}
