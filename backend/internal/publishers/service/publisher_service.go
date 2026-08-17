package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/publishers"
	"records-manager/backend/internal/publishers/repository"
)

type PublisherService struct {
	repo repository.PublisherRepository
}

func NewPublisherService(repo repository.PublisherRepository) *PublisherService {
	return &PublisherService{repo: repo}
}

func (s *PublisherService) CreatePublisher(ctx context.Context, name, description string) (*publishers.Publisher, error) {
	if name == "" {
		return nil, errors.New("le nom de l'éditeur est requis")
	}

	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("un éditeur avec ce nom existe déjà")
	}

	publisher := &publishers.Publisher{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, publisher)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

func (s *PublisherService) GetAllPublishers(ctx context.Context) ([]publishers.Publisher, error) {
	return s.repo.FindAll(ctx)
}

func (s *PublisherService) UpdatePublisher(ctx context.Context, id int, name, description string) (*publishers.Publisher, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("éditeur non trouvé")
	}

	if name != existing.Name {
		withSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if withSameName != nil && withSameName.ID != id {
			return nil, errors.New("un autre éditeur avec ce nom existe déjà")
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

func (s *PublisherService) DeletePublisher(ctx context.Context, id int) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("éditeur non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

func (s *PublisherService) GetPublisherByID(ctx context.Context, id int) (*publishers.Publisher, error) {
	publisher, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if publisher == nil {
		return nil, errors.New("éditeur non trouvé")
	}

	return publisher, nil
}

// CreatePublisherIfNotExists crée un éditeur seulement s'il n'existe pas déjà
func (s *PublisherService) CreatePublisherIfNotExists(ctx context.Context, name, description string) (*publishers.Publisher, error) {
	if name == "" {
		return nil, errors.New("le nom de l'éditeur est requis")
	}

	return s.repo.CreateIfNotExists(ctx, name, description)
}
