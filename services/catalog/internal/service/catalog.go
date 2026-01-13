package service

import (
	"catalog/internal/model"
	"catalog/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogService struct {
	db       *pgxpool.Pool
	versions repository.VersionRepository
	species  repository.SpeciesRepository
}

func NewCatalogService(db *pgxpool.Pool, v repository.VersionRepository, s repository.SpeciesRepository) *CatalogService {
	return &CatalogService{db: db, versions: v, species: s}
}

func (s *CatalogService) GetVersion(ctx context.Context) (int, error) {
	return s.versions.Get(ctx)
}

func (s *CatalogService) ListSpecies(ctx context.Context, activeOnly bool) ([]model.Species, error) {
	return s.species.List(ctx, activeOnly)
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
