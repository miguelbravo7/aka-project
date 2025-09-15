package repository

import (
	"context"
	"testing"

	"aka-project/internal/db"
	"aka-project/tests"

	"github.com/stretchr/testify/assert"
)

func TestCharacterRepo_GetCharacters_CreatesMissing(t *testing.T) {
	var createdCharacter db.Character

	mockQuerier := &tests.MockQueries{
		MissingIDsFunc: func(ctx context.Context, ids []int32) ([]int32, error) {
			return []int32{1}, nil
		},
		CreateCharacterFunc: func(ctx context.Context, arg db.CreateCharacterParams) (db.Character, error) {
			createdCharacter = db.Character{ID: 1, Name: arg.Name}
			return createdCharacter, nil
		},
	}

	repo := NewCharacterRepo(mockQuerier, tests.MockFetchOK)

	_, err := repo.GetCharacters(context.Background(), "", "", "")
	assert.NoError(t, err)

	assert.Equal(t, int32(1), createdCharacter.ID)
	assert.Equal(t, "Rick", createdCharacter.Name)
}
