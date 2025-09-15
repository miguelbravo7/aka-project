package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"aka-project/internal"
	"aka-project/internal/db"
	"aka-project/internal/repository"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type CharactersRepo interface {
	GetCharacters(ctx context.Context, species string, status string, origin string) (repository.CharactersResponse, error)
}

type CharacterHandler struct {
	Repo                       CharactersRepo
	Meter                      metric.Meter
	requestCounter             metric.Int64Counter
	errorCounter               metric.Int64Counter
	durationHistogram          metric.Int64Histogram
	charactersProcessedCounter metric.Int64Counter
}

func NewCharacterHandler(repo CharactersRepo, meter metric.Meter) (*CharacterHandler, error) {
	requestCounter, err := meter.Int64Counter(
		"api.characters.requests_total",
		metric.WithDescription("Total number of requests to the /characters endpoint"),
	)
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter(
		"api.characters.errors_total",
		metric.WithDescription("Total number of errors from the /characters endpoint"),
	)
	if err != nil {
		return nil, err
	}

	durationHistogram, err := meter.Int64Histogram(
		"api.characters.request_duration_seconds",
		metric.WithDescription("Duration of requests to the /characters endpoint in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	charactersProcessedCounter, err := meter.Int64Counter(
		"api.characters.processed_total",
		metric.WithDescription("Total number of characters processed by the /characters endpoint"),
	)
	if err != nil {
		return nil, err
	}

	return &CharacterHandler{
		Repo:                       repo,
		Meter:                      meter,
		requestCounter:             requestCounter,
		errorCounter:               errorCounter,
		durationHistogram:          durationHistogram,
		charactersProcessedCounter: charactersProcessedCounter,
	}, nil
}

type CharactersResponse struct {
	Info struct {
		Next  string `json:"next"`
		Prev  string `json:"prev"`
		Count int    `json:"count"`
		Pages int    `json:"pages"`
	} `json:"info"`
	Results []db.Character `json:"results"`
}

func (h *CharacterHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := log.Ctx(ctx)

	start := time.Now()
	h.requestCounter.Add(ctx, 1)

	characterResponse, err := h.Repo.GetCharacters(
		ctx,
		r.URL.Query().Get("species"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("origin"))

	if err != nil {
		h.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", err.Error())))
		log.Error().Err(err).Msg("failed to get characters")
		//nolint:errorlint
		switch e := err.(*internal.Error); e.Code {
		case internal.ErrorCodeNotFound:
			http.Error(w, e.Message, http.StatusNotFound)
		case internal.ErrorCodeUnauthorized:
			http.Error(w, e.Message, http.StatusUnauthorized)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	h.charactersProcessedCounter.Add(ctx, int64(len(characterResponse.Results)))

	duration := time.Since(start).Milliseconds()
	h.durationHistogram.Record(ctx, duration)

	writeJSON(w, characterResponse)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
