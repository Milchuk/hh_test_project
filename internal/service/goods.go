package service

import (
	"context"
	"errors"
	"fmt"
	"hh_test_project/internal/models"
	"hh_test_project/internal/repo"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNotFound         = errors.New("goods not found")
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidProjectID = errors.New("invalid project ID")
	ErrProjectNotExist  = errors.New("project does not exist")
)

type GoodsService struct {
	repo      *repo.GoodsRepository
	validator *validator.Validate
}

func NewGoodsService(repo *repo.GoodsRepository) *GoodsService {
	v := validator.New()
	v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return strings.TrimSpace(value) != ""
	})

	return &GoodsService{
		repo:      repo,
		validator: v,
	}
}

func (s *GoodsService) validate(input interface{}) error {
	if err := s.validator.Struct(input); err != nil {
		var valErr validator.ValidationErrors
		if errors.As(err, &valErr) {
			for _, e := range valErr {
				switch e.Tag() {
				case "required":
					return fmt.Errorf("%w: field '%s' is required", ErrValidationFailed, e.Field())
				case "notblank":
					return fmt.Errorf("%w: field '%s' cannot be blank", ErrValidationFailed, e.Field())
				case "max":
					return fmt.Errorf("%w: field '%s' exceeds max length (%s)", ErrValidationFailed, e.Field(), e.Param())
				}
			}
		}
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}
	return nil
}

func (s *GoodsService) GetByID(ctx context.Context, projectID, goodsID int) (*models.Goods, error) {
	if projectID <= 0 || goodsID <= 0 {
		return nil, ErrInvalidProjectID
	}

	goods, err := s.repo.GetByID(ctx, projectID, goodsID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get goods: %w", err)
	}

	return goods, nil
}

func (s *GoodsService) Create(ctx context.Context, projectID int, input models.GoodsCreate) (*models.Goods, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}

	exists, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project check failed: %w", err)
	}
	if !exists {
		return nil, ErrProjectNotExist
	}

	goods := models.GoodsCreate{
		Name: strings.TrimSpace(input.Name),
	}

	created, err := s.repo.Create(ctx, projectID, goods)
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}

	return created, nil
}

func (s *GoodsService) Update(ctx context.Context, projectID, goodsID int, input models.GoodsUpdate) (*models.Goods, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		*input.Description = strings.TrimSpace(*input.Description)
	}

	updated, err := s.repo.Update(ctx, projectID, goodsID, input)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update failed: %w", err)
	}

	return updated, nil
}

func (s *GoodsService) Delete(ctx context.Context, projectID, goodsID int) error {
	if err := s.repo.Delete(ctx, projectID, goodsID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}
