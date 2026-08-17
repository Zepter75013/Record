package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/gamegenres"
	"records-manager/backend/internal/gamegenres/repository"
)

type GameGenreService struct {
	repo repository.GameGenreRepository
}

func NewGameGenreService(repo repository.GameGenreRepository) *GameGenreService {
	return &GameGenreService{repo: repo}
}

func (s *GameGenreService) CreateGameGenre(ctx context.Context, name, description string) (*gamegenres.GameGenre, error) {
	if name == "" {
		return nil, errors.New("le nom du genre est requis")
	}

	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("un genre avec ce nom existe déjà")
	}

	genre := &gamegenres.GameGenre{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, genre)
	if err != nil {
		return nil, err
	}

	return genre, nil
}

func (s *GameGenreService) GetAllGameGenres(ctx context.Context) ([]gamegenres.GameGenre, error) {
	return s.repo.FindAll(ctx)
}

func (s *GameGenreService) UpdateGameGenre(ctx context.Context, id int, name, description string) (*gamegenres.GameGenre, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("genre non trouvé")
	}

	if name != existing.Name {
		withSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if withSameName != nil && withSameName.ID != id {
			return nil, errors.New("un autre genre avec ce nom existe déjà")
		}
	}

	existing.Name = name
	existing.Description = description

	err = s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *GameGenreService) DeleteGameGenre(ctx context.Context, id int) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("genre non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

func (s *GameGenreService) GetGameGenreByID(ctx context.Context, id int) (*gamegenres.GameGenre, error) {
	genre, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if genre == nil {
		return nil, errors.New("genre non trouvé")
	}

	return genre, nil
}

// CreateGameGenreIfNotExists crée un genre seulement s'il n'existe pas déjà
func (s *GameGenreService) CreateGameGenreIfNotExists(ctx context.Context, name, description string) (*gamegenres.GameGenre, error) {
	if name == "" {
		return nil, errors.New("le nom du genre est requis")
	}

	return s.repo.CreateIfNotExists(ctx, name, description)
}
