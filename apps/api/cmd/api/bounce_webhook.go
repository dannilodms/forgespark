package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	emailpostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/postgres"
	emailresend "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/resend"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/clock"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// mountBounceWebhook registers the webhook that takes the notices of a
// bounce Resend sends, when RESEND_WEBHOOK_SECRET is set. With no secret the
// path is unknown, as any other, so an installation without Resend serves
// nothing there.
//
// The webhook is a handler of its own on the mux, beside the REST adapter and
// not in it: the OpenAPI spec, which mountREST gives, leaves it out, and no
// middleware of the REST adapter runs on it, the check of the origin
// included. It carries no cookie: its signature is what authenticates it.
func mountBounceWebhook(mux *http.ServeMux, pool *pgxpool.Pool, settings config.Config) {
	if !settings.ResendWebhookSecret.IsSet() {
		return
	}
	markBounced := emailapplication.NewMarkEmailBounced(
		database.NewTransactions(pool),
		emailpostgres.NewEmailRepository(pool),
		emailpostgres.NewBounceNotices(pool),
		clock.System{},
	)
	emailresend.NewBounceWebhook(settings.ResendWebhookSecret, bodyLimits, markBounced, clock.System{}).Mount(mux)
}
