package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/genres"
	"records-manager/backend/internal/genres/repository"
)

type GenreService struct {
	repo repository.GenreRepository
}

func NewGenreService(repo repository.GenreRepository) *GenreService {
	return &GenreService{repo: repo}
}

func (s *GenreService) CreateGenre(ctx context.Context, name, description string) (*genres.Genre, error) {
	if name == "" {
		return nil, errors.New("le nom du genre est requis")
	}

	existingGenre, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existingGenre != nil {
		return nil, errors.New("un genre avec ce nom existe déjà")
	}

	genre := &genres.Genre{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, genre)
	if err != nil {
		return nil, err
	}

	return genre, nil
}

// CreateGenreIfNotExists - Crée un genre s'il n'existe pas déjà
func (s *GenreService) CreateGenreIfNotExists(ctx context.Context, name, description string) (*genres.Genre, error) {
	if name == "" {
		return nil, errors.New("le nom du genre est requis")
	}

	// Vérifier si le genre existe déjà
	existingGenre, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Si le genre existe déjà, le retourner
	if existingGenre != nil {
		return existingGenre, nil
	}

	// Sinon, créer le nouveau genre
	genre := &genres.Genre{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, genre)
	if err != nil {
		return nil, err
	}

	return genre, nil
}

func (s *GenreService) GetAllGenres(ctx context.Context) ([]genres.Genre, error) {
	return s.repo.FindAll(ctx)
}

func (s *GenreService) UpdateGenre(ctx context.Context, id int, name, description string) (*genres.Genre, error) {
	existingGenre, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingGenre == nil {
		return nil, errors.New("genre non trouvé")
	}

	if name != existingGenre.Name {
		genreWithSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if genreWithSameName != nil && genreWithSameName.ID != id {
			return nil, errors.New("un autre genre avec ce nom existe déjà")
		}
	}

	existingGenre.Name = name
	existingGenre.Description = description

	err = s.repo.Update(ctx, existingGenre)
	if err != nil {
		return nil, err
	}

	return existingGenre, nil
}

func (s *GenreService) DeleteGenre(ctx context.Context, id int) error {
	existingGenre, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingGenre == nil {
		return errors.New("genre non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

// GetGenreByID - Récupère un genre par son ID
func (s *GenreService) GetGenreByID(ctx context.Context, id int) (*genres.Genre, error) {
	genre, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if genre == nil {
		return nil, errors.New("genre non trouvé")
	}

	return genre, nil
}
