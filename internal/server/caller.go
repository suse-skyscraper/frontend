package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/suse-skyscraper/skyscraper/internal/application"
	"github.com/suse-skyscraper/skyscraper/internal/auth"
	"github.com/suse-skyscraper/skyscraper/internal/db"
	"github.com/suse-skyscraper/skyscraper/internal/server/middleware"
	"github.com/suse-skyscraper/skyscraper/internal/server/responses"
)

func V1CallerProfile(app *application.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := r.Context().Value(middleware.ContextCaller).(auth.Caller)
		if !ok {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		// Only show the profile for users
		if caller.Type != auth.CallerUser {
			_ = render.Render(w, r, responses.ErrNotFound)
			return
		}

		user, err := app.Repository.FindUser(r.Context(), caller.ID.String())
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		_ = render.Render(w, r, responses.NewUserResponse(user))
	}
}

func V1CallerCloudAccounts(app *application.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := r.Context().Value(middleware.ContextCaller).(auth.Caller)
		if !ok {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		organizationalUnits, err := callerOrganizationalUnits(r.Context(), app, caller)
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		ids := make([]uuid.UUID, 0, len(organizationalUnits))
		for _, ou := range organizationalUnits {
			ids = append(ids, ou.ID)
		}

		cloudAccounts, err := app.Repository.OrganizationalUnitsCloudAccounts(r.Context(), ids)
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		_ = render.Render(w, r, responses.NewCloudAccountListResponse(cloudAccounts))
	}
}

func callerOrganizationalUnits(ctx context.Context, app *application.App, caller auth.Caller) ([]db.OrganizationalUnit, error) {
	if caller.Type == auth.CallerUser {
		return app.Repository.GetUserOrganizationalUnits(ctx, caller.ID)
	} else if caller.Type == auth.CallerAPIKey {
		return app.Repository.GetAPIKeysOrganizationalUnits(ctx, caller.ID)
	}

	return nil, fmt.Errorf("caller not recognized")
}