package repository_test

import (
	"aka-project/internal/db"
	"aka-project/internal/repository"
	"aka-project/tests"
	"context"
	"reflect"
	"testing"
)

func TestMatchingIDs(t *testing.T) {
	chars := []db.Character{
		{ID: 1, Name: "Rick"},
		{ID: 2, Name: "Morty"},
		{ID: 3, Name: "Summer"},
	}

	ids := []int32{1, 3}

	want := []db.Character{
		{ID: 1, Name: "Rick"},
		{ID: 3, Name: "Summer"},
	}

	got := repository.MatchingIDs(ids, chars)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestGetCharacters_Success(t *testing.T) {
	mockQ := &tests.MockQueries{
		MissingIDsFunc: func(ctx context.Context, ids []int32) ([]int32, error) {
			return []int32{99}, nil
		},
		CreateCharacterFunc: func(ctx context.Context, arg db.CreateCharacterParams) (db.Character, error) {
			return db.Character{ID: 99, Name: arg.Name}, nil
		},
	}

	repo := repository.NewCharacterRepo((db.Querier)(nil), tests.MockFetchOK)
	repo.Queries = (db.Querier)(mockQ)

	resp, err := repo.GetCharacters(context.Background(), "Human", "Alive", "Earth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) == 0 || resp.Results[0].Name != "Rick" {
		t.Errorf("expected Rick, got %v", resp.Results)
	}
}

func TestGetCharacters_FetchError(t *testing.T) {
	repo := repository.NewCharacterRepo((*db.Queries)(nil), tests.MockFetchError)

	_, err := repo.GetCharacters(context.Background(), "Human", "Alive", "Earth")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
