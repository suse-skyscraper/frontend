package server

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/render"
	"github.com/suse-skyscraper/skyscraper/internal/application"
	"github.com/suse-skyscraper/skyscraper/internal/auth/apikeys"
	"github.com/suse-skyscraper/skyscraper/internal/db"
	"github.com/suse-skyscraper/skyscraper/internal/server/auditors"
	"github.com/suse-skyscraper/skyscraper/internal/server/middleware"
	"github.com/suse-skyscraper/skyscraper/internal/server/payloads"
	"github.com/suse-skyscraper/skyscraper/internal/server/responses"
)

func V1ListAPIKeys(app *application.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKeys, err := app.Repository.GetAPIKeys(r.Context())
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		_ = render.Render(w, r, responses.NewAPIKeysResponse(apiKeys))
	}
}

func V1GetAPIKey(_ *application.App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, ok := r.Context().Value(middleware.ContextAPIKey).(db.ApiKey)
		if !ok {
			_ = render.Render(w, r, responses.ErrNotFound)
			return
		}

		_ = render.Render(w, r, responses.NewAPIKeyResponse(apiKey, ""))
	}
}

func V1CreateAPIKey(app *application.App) func(w http.ResponseWriter, r *http.Request) {
	apiKeyGenerator := apikeys.NewGenerator(app)

	return func(w http.ResponseWriter, r *http.Request) {
		// bind the payload
		var payload payloads.CreateAPIKeyPayload
		err := render.Bind(r, &payload)
		if err != nil {
			_ = render.Render(w, r, responses.ErrInvalidRequest(err))
			return
		}

		// Begin a database transaction
		repo, err := app.Repository.Begin(r.Context())
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		// If any error occurs, rollback the transaction
		defer func(repo db.RepositoryQueries, ctx context.Context) {
			_ = repo.Rollback(ctx)
		}(repo, r.Context())

		// create an auditor within our transaction
		auditor := auditors.NewAuditor(repo)

		// create an API key
		encodedHash, token, err := apiKeyGenerator.Generate()
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		// persist the API key
		apiKey, err := app.Repository.CreateAPIKey(r.Context(), db.InsertAPIKeyParams{
			Owner:       payload.Data.Owner,
			Description: sql.NullString{String: payload.Data.Description, Valid: true},
			System:      false,
			Encodedhash: encodedHash,
		})
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		// audit the change
		err = auditor.Audit(r.Context(), db.AuditResourceTypeApiKey, apiKey.ID, payload)
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		// Commit the transaction
		err = repo.Commit(r.Context())
		if err != nil {
			_ = render.Render(w, r, responses.ErrInternalServerError)
			return
		}

		_ = render.Render(w, r, responses.NewAPIKeyResponse(apiKey, token))
	}
}