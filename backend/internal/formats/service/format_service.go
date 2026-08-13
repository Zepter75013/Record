package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/formats"
	"records-manager/backend/internal/formats/repository"
)

type FormatService struct {
	repo repository.FormatRepository
}

func NewFormatService(repo repository.FormatRepository) *FormatService {
	return &FormatService{repo: repo}
}

func (s *FormatService) CreateFormat(ctx context.Context, name, description string) (*formats.Format, error) {
	if name == "" {
		return nil, errors.New("le nom du format est requis")
	}

	existingFormat, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existingFormat != nil {
		return nil, errors.New("un format avec ce nom existe déjà")
	}

	format := &formats.Format{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, format)
	if err != nil {
		return nil, err
	}

	return format, nil
}

func (s *FormatService) GetAllFormats(ctx context.Context) ([]formats.Format, error) {
	return s.repo.FindAll(ctx)
}

func (s *FormatService) UpdateFormat(ctx context.Context, id int, name, description string) (*formats.Format, error) {
	existingFormat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingFormat == nil {
		return nil, errors.New("format non trouvé")
	}

	if name != existingFormat.Name {
		formatWithSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if formatWithSameName != nil && formatWithSameName.ID != id {
			return nil, errors.New("un autre format avec ce nom existe déjà")
		}
	}

	existingFormat.Name = name
	existingFormat.Description = description

	err = s.repo.Update(ctx, existingFormat)
	if err != nil {
		return nil, err
	}

	return existingFormat, nil
}

func (s *FormatService) DeleteFormat(ctx context.Context, id int) error {
	existingFormat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingFormat == nil {
		return errors.New("format non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

// GetFormatByID - Récupère un format par son ID
func (s *FormatService) GetFormatByID(ctx context.Context, id int) (*formats.Format, error) {
	format, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if format == nil {
		return nil, errors.New("format non trouvé")
	}

	return format, nil
}

// CreateFormatIfNotExists crée un format seulement s'il n'existe pas déjà
func (s *FormatService) CreateFormatIfNotExists(ctx context.Context, name, description string) (*formats.Format, error) {
	if name == "" {
		return nil, errors.New("le nom du format est requis")
	}

	return s.repo.CreateIfNotExists(ctx, name, description)
}
