package tests

import (
	"aka-project/internal/db"
	"aka-project/internal/helper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func MockFetchOK(ctx context.Context, url string) (*helper.APIResponse, error) {
	return &helper.APIResponse{
		Info: struct {
			Next  string `json:"next"`
			Prev  string `json:"prev"`
			Count int    `json:"count"`
			Pages int    `json:"pages"`
		}{Next: "next", Prev: "prev", Count: 1, Pages: 1},
		Results: []json.RawMessage{[]byte(fmt.Sprintf(`
		{
			"id": 1,
			"name": "Rick",
			"status": "Alive",
			"species": "Human",
			"gender": "Male",
			"url": "http://example.com/rick",
			"created": "%s"
		}`, time.Now().Format(time.RFC3339)))},
	}, nil
}

func MockFetchError(ctx context.Context, url string) (*helper.APIResponse, error) {
	return nil, errors.New("fetch failed")
}

// MockQueries implements only the methods we need
type MockQueries struct {
	MissingIDsFunc      func(ctx context.Context, ids []int32) ([]int32, error)
	CreateCharacterFunc func(ctx context.Context, arg db.CreateCharacterParams) (db.Character, error)
}

func (m *MockQueries) GetMissingCharacterIDs(ctx context.Context, ids []int32) ([]int32, error) {
	return m.MissingIDsFunc(ctx, ids)
}

func (m *MockQueries) CreateCharacter(ctx context.Context, arg db.CreateCharacterParams) (db.Character, error) {
	return m.CreateCharacterFunc(ctx, arg)
}
