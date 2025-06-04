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
	app *app.App,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("PlanScheduler: context cancelled, stopping scheduler")
				return

			case now := <-ticker.C:
				duePlans, err := app.Plan.GetDueAndMark(ctx, now)
				if err != nil {
					log.Printf("PlanScheduler: failed to fetch due plans: %v", err)
					continue
				}
				for _, p := range duePlans {
					handlers.PlansRemindHandler(&p, app)(ctx, b, nil)
				}
			}
		}
	}()
}

func StartCatScheduler(
	ctx context.Context,
	b *bot.Bot,
	app *app.App,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("CatScheduler: context cancelled, stopping scheduler")
				return

			case now := <-ticker.C:
				dueUsers, err := app.User.FetchDueCats(ctx, now.UTC())
				if err != nil {
					log.Printf("CatScheduler: failed to fetch due cats: %v", err)
					continue
				}
				for _, u := range dueUsers {
					handlers.ScheduledComplimentImageHandler(ctx, app, b, u.ChatID)
				}
			}
		}
	}()
}
