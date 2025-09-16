package repository

import (
	"context"
	"encoding/json"
	"net/url"

	"aka-project/internal"
	"aka-project/internal/config"
	"aka-project/internal/db"
	"aka-project/internal/helper"

	"github.com/rs/zerolog/log"
)

var RM_API_ENDPOINT string = config.Load().RMAPI

type CharacterRepo struct {
	Queries db.Querier
	Fetch   func(ctx context.Context, url string) (*helper.APIResponse, error)
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

func NewCharacterRepo(queries db.Querier, fetch func(ctx context.Context, url string) (*helper.APIResponse, error)) *CharacterRepo {
	return &CharacterRepo{
		Queries: queries,
		Fetch:   fetch,
	}
}

func (repo *CharacterRepo) GetCharacters(ctx context.Context, species string, status string, origin string) (CharactersResponse, error) {
	url, err := url.Parse(RM_API_ENDPOINT)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse URL")
		return CharactersResponse{}, internal.NewError(internal.ErrorCodeInternal, "failed to parse URL")
	}

	query := url.Query()
	query.Set("species", species)
	query.Set("status", status)
	query.Set("origin", origin)
	url.RawQuery = query.Encode()

	resp, err := repo.Fetch(ctx, url.String())
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch characters")
		return CharactersResponse{}, internal.NewError(internal.ErrorCodeInternal, "failed to fetch characters")
	}

	result := CharactersResponse{
		Info:    resp.Info,
		Results: []db.Character{},
	}

	for _, rawChar := range resp.Results {
		var char db.Character
		if err := json.Unmarshal(rawChar, &char); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal character")
			continue
		}
		result.Results = append(result.Results, char)
	}

	err = repo.persistCharacters(result.Results)
	if err != nil {
		return CharactersResponse{}, err
	}

	return result, nil
}

func (repo *CharacterRepo) persistCharacters(characters []db.Character) error {
	missingCharacters, err := repo.getMissingCharacters(characters)
	if err != nil {
		return internal.Wrap(err, internal.NewError(internal.ErrorCodeInternal, "failed to get missing characters"))
	}

	for _, character := range missingCharacters {
		_, err = repo.Queries.CreateCharacter(context.Background(), db.CreateCharacterParams(character))
		if err != nil {
			return internal.Wrap(err, internal.NewError(internal.ErrorCodeInternal, "failed to create character"))
		}
	}

	return nil
}

func (repo *CharacterRepo) getMissingCharacters(characters []db.Character) ([]db.Character, error) {
	var ids []int32
	for _, character := range characters {
		ids = append(ids, character.ID)
	}

	missingIDs, err := repo.Queries.GetMissingCharacterIDs(context.Background(), ids)
	if err != nil {
		return nil, internal.Wrap(err, internal.NewError(internal.ErrorCodeInternal, "failed to get missing character IDs"))
	}

	return MatchingIDs(missingIDs, characters), nil
}

func MatchingIDs(ids []int32, characters []db.Character) []db.Character {
	// Build a set from ids
	idSet := make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	// Collect matches
	var matches []db.Character
	for _, c := range characters {
		if _, exists := idSet[c.ID]; exists {
			matches = append(matches, c)
		}
	}
	return matches
}
