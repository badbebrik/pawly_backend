module chat

go 1.24.7

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	pawly/pkg v0.0.0
)

replace pawly/pkg => ../pkg
