package repopg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

const (
	pgUniqueViolation    = "23505" // UNIQUE constraint violation
	pgExclusionViolation = "23P01" // EXCLUDE USING gist violation (solapamiento de reservas)
	pgInvalidTextRepr    = "22P02" // invalid_text_representation: valor no casteable al tipo de columna (e.g. "no-existe" en columna uuid)
)

var _ ports.UserRepository = (*UserRepoPg)(nil)

type UserRepoPg struct {
	pool *pgxpool.Pool
}

func NewUserRepoPg(pool *pgxpool.Pool) *UserRepoPg {
	return &UserRepoPg{pool: pool}
}

func (r *UserRepoPg) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users WHERE id = $1`

	return queryOneUser(ctx, r.pool, q, id)
}

func (r *UserRepoPg) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users WHERE email = $1`

	return queryOneUser(ctx, r.pool, q, email)
}

// Create inserta un usuario. El ID debe venir generado desde la capa de servicio.
// Retorna domain.ErrConflict si el email ya existe (unique violation).
func (r *UserRepoPg) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, q, u.ID, u.Name, u.Email, u.Password, u.Role).
		Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return pgErr(err)
	}
	return nil
}

// Update modifica name y/o email. Campos vacíos se ignoran via COALESCE+NULLIF.
func (r *UserRepoPg) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	const q = `
		UPDATE users
		SET
			name       = COALESCE(NULLIF($2, ''), name),
			email      = COALESCE(NULLIF($3, ''), email),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, password, role, created_at, updated_at`

	return queryOneUser(ctx, r.pool, q, id, req.Name, req.Email)
}

func (r *UserRepoPg) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// queryOneUser ejecuta q y escanea la fila resultante usando pgx.RowToStructByName,
// que mapea columnas a campos del struct por el tag db:"...".
func queryOneUser(ctx context.Context, pool *pgxpool.Pool, q string, args ...any) (*domain.User, error) {
	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect user: %w", err)
	}
	return &u, nil
}

// pgErr traduce errores de pgconn a errores de dominio para que
// la service layer no necesite importar pgx.
func pgErr(err error) error {
	var e *pgconn.PgError
	if errors.As(err, &e) {
		switch e.Code {
		case pgUniqueViolation, pgExclusionViolation:
			return domain.ErrConflict
		case pgInvalidTextRepr:
			// El valor no es casteable al tipo de la columna (e.g. string no-UUID en columna uuid).
			// Se trata como "no encontrado" para no exponer detalles del schema al cliente.
			return domain.ErrNotFound
		}
	}
	return fmt.Errorf("db: %w", err)
}
