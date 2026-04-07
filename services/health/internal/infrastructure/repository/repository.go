package repository

import (
	repo "health/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	*LogRepository
	*ScheduledRepository
}

func NewRepository(db *pgxpool.Pool) repo.Repository {
	return &Repository{
		LogRepository:       NewLogRepository(db),
		ScheduledRepository: NewScheduledRepository(db),
	}
}

var _ repo.Repository = (*Repository)(nil)
