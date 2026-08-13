// records-manager/backend/internal/labels/service/label_service.go
package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/labels"
	"records-manager/backend/internal/labels/repository"
)

type LabelService struct {
	repo repository.LabelRepository
}

func NewLabelService(repo repository.LabelRepository) *LabelService {
	return &LabelService{repo: repo}
}

func (s *LabelService) CreateLabel(ctx context.Context, name, description string) (*labels.Label, error) {
	if name == "" {
		return nil, errors.New("le nom du label est requis")
	}

	existingLabel, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if existingLabel != nil {
		return nil, errors.New("un label avec ce nom existe déjà")
	}

	label := &labels.Label{
		Name:        name,
		Description: description,
	}

	err = s.repo.Create(ctx, label)
	if err != nil {
		return nil, err
	}

	return label, nil
}

func (s *LabelService) GetAllLabels(ctx context.Context) ([]labels.Label, error) {
	return s.repo.FindAll(ctx)
}

func (s *LabelService) UpdateLabel(ctx context.Context, id int, name, description string) (*labels.Label, error) {
	existingLabel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingLabel == nil {
		return nil, errors.New("label non trouvé")
	}

	if name != existingLabel.Name {
		labelWithSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}

		if labelWithSameName != nil && labelWithSameName.ID != id {
			return nil, errors.New("un autre label avec ce nom existe déjà")
		}
	}

	existingLabel.Name = name
	existingLabel.Description = description

	err = s.repo.Update(ctx, existingLabel)
	if err != nil {
		return nil, err
	}

	return existingLabel, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, id int) error {
	existingLabel, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingLabel == nil {
		return errors.New("label non trouvé")
	}

	return s.repo.Delete(ctx, id)
}

// GetLabelByID - Récupère un label par son ID
func (s *LabelService) GetLabelByID(ctx context.Context, id int) (*labels.Label, error) {
	label, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if label == nil {
		return nil, errors.New("label non trouvé")
	}

	return label, nil
}
