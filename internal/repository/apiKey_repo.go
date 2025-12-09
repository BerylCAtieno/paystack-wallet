package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
)

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(key *auth.APIKey) error {
	permsJSON, err := json.Marshal(key.Permissions)
	if err != nil {
		return err
	}

	query := `INSERT INTO api_keys (id, user_id, name, key_hash, permissions, expires_at, is_revoked, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.Exec(query,
		key.ID, key.UserID, key.Name, key.KeyHash, string(permsJSON),
		key.ExpiresAt, key.IsRevoked, key.CreatedAt, key.UpdatedAt,
	)
	return err
}

func (r *APIKeyRepository) GetByID(id string) (*auth.APIKey, error) {
	query := `SELECT id, user_id, name, key_hash, permissions, expires_at, is_revoked, created_at, updated_at
		FROM api_keys WHERE id = ?`

	key := &auth.APIKey{}
	var permsJSON string

	err := r.db.QueryRow(query, id).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &permsJSON,
		&key.ExpiresAt, &key.IsRevoked, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(permsJSON), &key.Permissions); err != nil {
		return nil, err
	}

	return key, nil
}

func (r *APIKeyRepository) GetByKeyHash(hash string) (*auth.APIKey, error) {
	query := `SELECT id, user_id, name, key_hash, permissions, expires_at, is_revoked, created_at, updated_at
		FROM api_keys WHERE key_hash = ?`

	key := &auth.APIKey{}
	var permsJSON string

	err := r.db.QueryRow(query, hash).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyHash, &permsJSON,
		&key.ExpiresAt, &key.IsRevoked, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(permsJSON), &key.Permissions); err != nil {
		return nil, err
	}

	return key, nil
}

func (r *APIKeyRepository) CountActiveByUserID(userID string) (int, error) {
	query := `SELECT COUNT(*) FROM api_keys 
		WHERE user_id = ? AND is_revoked = 0 AND expires_at > CURRENT_TIMESTAMP`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *APIKeyRepository) Update(key *auth.APIKey) error {
	permsJSON, err := json.Marshal(key.Permissions)
	if err != nil {
		return err
	}

	query := `UPDATE api_keys SET is_revoked = ?, permissions = ?, updated_at = ? WHERE id = ?`

	_, err = r.db.Exec(query, key.IsRevoked, string(permsJSON), key.UpdatedAt, key.ID)
	return err
}

func (r *APIKeyRepository) ListByUserID(userID string) ([]*auth.APIKey, error) {
	query := `SELECT id, user_id, name, key_hash, permissions, expires_at, is_revoked, created_at, updated_at
		FROM api_keys WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*auth.APIKey
	for rows.Next() {
		key := &auth.APIKey{}
		var permsJSON string

		err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyHash, &permsJSON,
			&key.ExpiresAt, &key.IsRevoked, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(permsJSON), &key.Permissions); err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return keys, nil
}
