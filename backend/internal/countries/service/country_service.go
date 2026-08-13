package service

import (
	"context"
	"errors"
	"records-manager/backend/internal/countries"
	"records-manager/backend/internal/countries/repository"
)

type CountryService struct {
	repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) *CountryService {
	return &CountryService{repo: repo}
}

func (s *CountryService) CreateCountry(ctx context.Context, name, code, description string) (*countries.Country, error) {
	if name == "" {
		return nil, errors.New("le nom du pays est requis")
	}
	if code == "" {
		return nil, errors.New("le code du pays est requis")
	}

	// Vérifier si le pays existe déjà par nom
	existingByName, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existingByName != nil {
		return nil, errors.New("un pays avec ce nom existe déjà")
	}

	// Vérifier si le pays existe déjà par code
	existingByCode, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existingByCode != nil {
		return nil, errors.New("un pays avec ce code existe déjà")
	}

	country := &countries.Country{
		Name:        name,
		Code:        code,
		Description: description,
	}

	err = s.repo.Create(ctx, country)
	if err != nil {
		return nil, err
	}

	return country, nil
}

func (s *CountryService) GetAllCountries(ctx context.Context) ([]countries.Country, error) {
	return s.repo.FindAll(ctx)
}

func (s *CountryService) GetCountryByID(ctx context.Context, id int) (*countries.Country, error) {
	country, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if country == nil {
		return nil, errors.New("pays non trouvé")
	}

	return country, nil
}

func (s *CountryService) UpdateCountry(ctx context.Context, id int, name, code, description string) (*countries.Country, error) {
	existingCountry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingCountry == nil {
		return nil, errors.New("pays non trouvé")
	}

	// Vérifier les conflits de nom
	if name != existingCountry.Name {
		countryWithSameName, err := s.repo.FindByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if countryWithSameName != nil && countryWithSameName.ID != id {
			return nil, errors.New("un autre pays avec ce nom existe déjà")
		}
	}

	// Vérifier les conflits de code
	if code != existingCountry.Code {
		countryWithSameCode, err := s.repo.FindByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if countryWithSameCode != nil && countryWithSameCode.ID != id {
			return nil, errors.New("un autre pays avec ce code existe déjà")
		}
	}

	existingCountry.Name = name
	existingCountry.Code = code
	existingCountry.Description = description

	err = s.repo.Update(ctx, existingCountry)
	if err != nil {
		return nil, err
	}

	return existingCountry, nil
}

func (s *CountryService) DeleteCountry(ctx context.Context, id int) error {
	existingCountry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingCountry == nil {
		return errors.New("pays non trouvé")
	}

	return s.repo.Delete(ctx, id)
}
