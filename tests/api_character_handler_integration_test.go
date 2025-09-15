package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aka-project/internal"
	"aka-project/internal/api"
	"aka-project/internal/db"
	"aka-project/internal/repository"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric/noop"
)

// Fake repo implements CharacterRepo interface
type fakeCharacterRepo struct {
	users       []db.Character
	returnError bool
}

func (f *fakeCharacterRepo) GetCharacters(ctx context.Context, species string, status string, origin string) (repository.CharactersResponse, error) {
	if f.returnError {
		return repository.CharactersResponse{}, internal.NewError(internal.ErrorCodeInternal, "something went wrong")
	}
	return repository.CharactersResponse{
		Info: struct {
			Next  string `json:"next"`
			Prev  string `json:"prev"`
			Count int    `json:"count"`
			Pages int    `json:"pages"`
		}{Next: "", Prev: "", Count: len(f.users), Pages: 1},
		Results: f.users,
	}, nil
}

func TestCreateCharactersHandler(t *testing.T) {
	// Prepare fake users
	createdAt := time.Now().UTC()
	users := []db.Character{
		{ID: 1, Name: "Alice", Species: "alien", Created: createdAt},
		{ID: 2, Name: "Bob", Species: "human", Created: createdAt.Add(1 * time.Minute)},
		{ID: 3, Name: "Charlie", Species: "human", Created: createdAt.Add(2 * time.Minute)},
	}

	// Inject fake repo into handler
	repo := &fakeCharacterRepo{
		users: users,
	}
	handler, err := api.NewCharacterHandler(repo, noop.NewMeterProvider().Meter("test"))
	assert.NoError(t, err)

	// Create test HTTP request
	req := httptest.NewRequest("GET", "/characters", nil)
	w := httptest.NewRecorder()

	// Call the actual CreateCharacters method
	handler.GetCharacters(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var body api.CharactersResponse

	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Len(t, body.Results, len(users))
	assert.Equal(t, body.Info.Count, len(users))

	assert.Equal(t, users[0].ID, body.Results[0].ID)
	assert.Equal(t, users[1].ID, body.Results[1].ID)
	assert.Equal(t, users[0].Created.Format(time.RFC3339), body.Results[0].Created.Format(time.RFC3339))
	assert.Equal(t, users[1].Created.Format(time.RFC3339), body.Results[1].Created.Format(time.RFC3339))
}

func TestCreateCharactersHandler_Error(t *testing.T) {
	// Inject fake repo into handler
	repo := &fakeCharacterRepo{
		returnError: true,
	}
	handler, err := api.NewCharacterHandler(repo, noop.NewMeterProvider().Meter("test"))
	assert.NoError(t, err)

	// Create test HTTP request
	req := httptest.NewRequest("GET", "/characters", nil)
	w := httptest.NewRecorder()

	// Call the actual CreateCharacters method
	handler.GetCharacters(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
