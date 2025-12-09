package auth

import (
	"errors"
	"time"

	"github.com/BerylCAtieno/paystack-wallet/internal/security"
)

var (
	ErrMaxAPIKeysReached  = errors.New("maximum of 5 active API keys allowed")
	ErrInvalidPermissions = errors.New("invalid permissions")
	ErrKeyNotExpired      = errors.New("key is not expired")
	ErrKeyNotFound        = errors.New("api key not found")
	ErrInvalidExpiry      = errors.New("invalid expiry duration")
)

type Service struct {
	repo APIKeyRepository
}

func NewService(repo APIKeyRepository) *Service {
	return &Service{repo: repo}
}

type APIKeyRepository interface {
	Create(key *APIKey) error
	GetByID(id string) (*APIKey, error)
	GetByKeyHash(hash string) (*APIKey, error)
	CountActiveByUserID(userID string) (int, error)
	Update(key *APIKey) error
	ListByUserID(userID string) ([]*APIKey, error)
}

func (s *Service) CreateAPIKey(userID, name string, permissions []Permission, expiry ExpiryDuration) (*APIKey, string, error) {
	// Validate permissions
	for _, perm := range permissions {
		if !ValidPermissions[perm] {
			return nil, "", ErrInvalidPermissions
		}
	}

	// Check active keys count
	count, err := s.repo.CountActiveByUserID(userID)
	if err != nil {
		return nil, "", err
	}
	if count >= 5 {
		return nil, "", ErrMaxAPIKeysReached
	}

	// Generate API key
	rawKey := security.GenerateAPIKey()
	keyHash := security.HashAPIKey(rawKey)

	apiKey := &APIKey{
		ID:          security.GenerateID(),
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHash,
		Permissions: permissions,
		ExpiresAt:   time.Now().Add(expiry.ToDuration()),
		IsRevoked:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *Service) RolloverAPIKey(userID, expiredKeyID string, newExpiry ExpiryDuration) (*APIKey, string, error) {
	oldKey, err := s.repo.GetByID(expiredKeyID)
	if err != nil {
		return nil, "", err
	}

	if oldKey.UserID != userID {
		return nil, "", ErrKeyNotFound
	}

	if !oldKey.IsExpired() {
		return nil, "", ErrKeyNotExpired
	}

	// Create new key with same permissions
	return s.CreateAPIKey(userID, oldKey.Name, oldKey.Permissions, newExpiry)
}

func (s *Service) ValidateAPIKey(rawKey string) (*APIKey, error) {
	keyHash := security.HashAPIKey(rawKey)
	apiKey, err := s.repo.GetByKeyHash(keyHash)
	if err != nil {
		return nil, err
	}

	if apiKey.IsRevoked {
		return nil, errors.New("api key is revoked")
	}

	if apiKey.IsExpired() {
		return nil, errors.New("api key is expired")
	}

	return apiKey, nil
}

func (k *APIKey) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}

func (k *APIKey) HasPermission(perm Permission) bool {
	for _, p := range k.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}
