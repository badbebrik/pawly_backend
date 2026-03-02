package service

import (
	"catalog/internal/model"
	"catalog/internal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogService struct {
	db       *pgxpool.Pool
	versions repository.VersionRepository

	breeds   repository.BreedRepository
	species  repository.SpeciesRepository
	colors   repository.ColorRepository
	patterns repository.PatternRepository
}

func NewCatalogService(
	db *pgxpool.Pool,
	v repository.VersionRepository,
	b repository.BreedRepository,
	s repository.SpeciesRepository,
	c repository.ColorRepository,
	p repository.PatternRepository,
) *CatalogService {
	return &CatalogService{
		db:       db,
		versions: v,
		breeds:   b,
		species:  s,
		colors:   c,
		patterns: p,
	}
}

func (s *CatalogService) GetVersion(ctx context.Context) (int, error) {
	return s.versions.Get(ctx)
}

func (s *CatalogService) ListSpecies(ctx context.Context, activeOnly bool) ([]model.Species, error) {
	return s.species.List(ctx, activeOnly)
}

func (s *CatalogService) ListBreeds(ctx context.Context, speciesID *uuid.UUID, activeOnly bool) ([]model.Breed, error) {
	return s.breeds.List(ctx, speciesID, activeOnly)
}

func (s *CatalogService) ListColors(ctx context.Context, activeOnly bool) ([]model.Color, error) {
	return s.colors.List(ctx, activeOnly)
}

func (s *CatalogService) ListPatterns(ctx context.Context, activeOnly bool) ([]model.Pattern, error) {
	return s.patterns.List(ctx, activeOnly)
}

func (s *CatalogService) AdminCreateSpecies(ctx context.Context, sp *model.Species) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	sp.Version = v

	if err := s.species.CreateTx(ctx, tx, sp); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}

func (s *CatalogService) AdminUpdateSpecies(ctx context.Context, sp *model.Species) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	sp.Version = v

	if err := s.species.UpdateTx(ctx, tx, sp); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}

func (s *CatalogService) AdminCreateColor(ctx context.Context, c *model.Color) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	c.Version = v

	if err := s.colors.CreateTx(ctx, tx, c); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}

func (s *CatalogService) AdminUpdateColor(ctx context.Context, c *model.Color) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	c.Version = v

	if err := s.colors.UpdateTx(ctx, tx, c); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}

func (s *CatalogService) AdminCreatePattern(ctx context.Context, ptn *model.Pattern) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	ptn.Version = v

	if err := s.patterns.CreateTx(ctx, tx, ptn); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}

func (s *CatalogService) AdminUpdatePattern(ctx context.Context, ptn *model.Pattern) (newCatalogVersion int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	v, err := s.versions.BumpTx(ctx, tx)
	if err != nil {
		return 0, err
	}
	ptn.Version = v

	if err := s.patterns.UpdateTx(ctx, tx, ptn); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return v, nil
}
