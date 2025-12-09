package repository

import (
	"database/sql"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *user.User) error {
	query := `INSERT INTO users (id, email, name, google_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, u.ID, u.Email, u.Name, u.GoogleID, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepository) GetByID(id string) (*user.User, error) {
	query := `SELECT id, email, name, google_id, created_at, updated_at FROM users WHERE id = ?`

	u := &user.User{}
	err := r.db.QueryRow(query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByGoogleID(googleID string) (*user.User, error) {
	query := `SELECT id, email, name, google_id, created_at, updated_at FROM users WHERE google_id = ?`

	u := &user.User{}
	err := r.db.QueryRow(query, googleID).Scan(
		&u.ID, &u.Email, &u.Name, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(email string) (*user.User, error) {
	query := `SELECT id, email, name, google_id, created_at, updated_at FROM users WHERE email = ?`

	u := &user.User{}
	err := r.db.QueryRow(query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.GoogleID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
