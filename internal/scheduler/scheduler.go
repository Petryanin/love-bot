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
				log.Print("PlanScheduler: context cancelled, stopping scheduler")
				return

			case now := <-ticker.C:
				duePlans, err := app.Plan.GetDueAndMark(ctx, now)
				if err != nil {
					log.Print("PlanScheduler: failed to fetch due plans: %w", err)
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
				log.Print("CatScheduler: context cancelled, stopping scheduler")
				return

			case now := <-ticker.C:
				dueUsers, err := app.User.FetchDueCats(ctx, now.UTC())
				if err != nil {
					log.Print("CatScheduler: failed to fetch due cats: %w", err)
					continue
				}
				for _, u := range dueUsers {
					handlers.ScheduledComplimentImageHandler(ctx, app, b, u.ChatID)
				}
			}
		}
	}()
}

func StartCleanupScheduler(
	ctx context.Context,
	app *app.App,
	interval time.Duration,
	retention time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Print("CleanupScheduler: context cancelled, stopping scheduler")
				return

			case <-ticker.C:
				removed, err := app.Plan.DeleteExpired(ctx, retention)
				if err != nil {
					log.Print("CleanupScheduler: error removing expired plans: %w", err)
				} else if removed > 0 {
					log.Printf("CleanupScheduler: removed %d expired plans", removed)
				}
			}
		}
	}()
}
