package service

import (
	"context"
	"fmt"
	"records-manager/backend/internal/artists"
	"records-manager/backend/internal/artists/repository"
)

type ArtistService struct {
	repo repository.ArtistRepository
}

func NewArtistService(repo repository.ArtistRepository) *ArtistService {
	return &ArtistService{repo: repo}
}

func (s *ArtistService) CreateArtist(ctx context.Context, name, biography string, countryID *int) (*artists.Artist, error) {
	// Vérifier si l'artiste existe déjà
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("un artiste avec ce nom existe déjà")
	}

	artist := &artists.Artist{
		Name:      name,
		Biography: biography,
		CountryID: countryID,
	}

	if err := s.repo.Create(ctx, artist); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, artist.ID)
}

func (s *ArtistService) GetAllArtists(ctx context.Context) ([]artists.Artist, error) {
	return s.repo.FindAll(ctx)
}

func (s *ArtistService) GetArtistByID(ctx context.Context, id int) (*artists.Artist, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ArtistService) UpdateArtist(ctx context.Context, id int, name, biography string, countryID *int) (*artists.Artist, error) {
	// Vérifier si un autre artiste avec ce nom existe
	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("un autre artiste avec ce nom existe déjà")
	}

	artist, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	artist.Name = name
	artist.Biography = biography
	artist.CountryID = countryID

	if err := s.repo.Update(ctx, artist); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *ArtistService) DeleteArtist(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
