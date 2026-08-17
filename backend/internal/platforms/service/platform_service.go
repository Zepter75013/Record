package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/platforms"
	"records-manager/backend/internal/platforms/repository"
)

type PlatformService struct {
	repo repository.PlatformRepository
}

func NewPlatformService(repo repository.PlatformRepository) *PlatformService {
	return &PlatformService{repo: repo}
}

func (s *PlatformService) CreatePlatform(ctx context.Context, name, description string) (*platforms.Platform, error) {
	if name == "" {
		return nil, errors.New("le nom de la plateforme est requis")
	}

	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("une plateforme avec ce nom existe déjà")
	}

	platform := &platforms.Platform{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, platform)
	if err != nil {
		return nil, err
	}

	return platform, nil
}

func (s *PlatformService) GetAllPlatforms(ctx context.Context) ([]platforms.Platform, error) {
	return s.repo.FindAll(ctx)
}

func (s *PlatformService) UpdatePlatform(ctx context.Context, id int, name, description string) (*platforms.Platform, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("plateforme non trouvée")
	}

	if name != existing.Name {
		withSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if withSameName != nil && withSameName.ID != id {
			return nil, errors.New("une autre plateforme avec ce nom existe déjà")
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

func (s *PlatformService) DeletePlatform(ctx context.Context, id int) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("plateforme non trouvée")
	}

	return s.repo.Delete(ctx, id)
}

func (s *PlatformService) GetPlatformByID(ctx context.Context, id int) (*platforms.Platform, error) {
	platform, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if platform == nil {
		return nil, errors.New("plateforme non trouvée")
	}

	return platform, nil
}

// CreatePlatformIfNotExists crée une plateforme seulement si elle n'existe pas déjà
func (s *PlatformService) CreatePlatformIfNotExists(ctx context.Context, name, description string) (*platforms.Platform, error) {
	if name == "" {
		return nil, errors.New("le nom de la plateforme est requis")
	}

	return s.repo.CreateIfNotExists(ctx, name, description)
}
