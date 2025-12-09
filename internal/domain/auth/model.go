package auth

import "time"

type APIKey struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Name        string       `json:"name"`
	KeyHash     string       `json:"-"`
	Permissions []Permission `json:"permissions"`
	ExpiresAt   time.Time    `json:"expires_at"`
	IsRevoked   bool         `json:"is_revoked"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission string

const (
	PermissionDeposit  Permission = "deposit"
	PermissionTransfer Permission = "transfer"
	PermissionRead     Permission = "read"
)

var ValidPermissions = map[Permission]bool{
	PermissionDeposit:  true,
	PermissionTransfer: true,
	PermissionRead:     true,
}

type ExpiryDuration string

const (
	Expiry1Hour  ExpiryDuration = "1H"
	Expiry1Day   ExpiryDuration = "1D"
	Expiry1Month ExpiryDuration = "1M"
	Expiry1Year  ExpiryDuration = "1Y"
)

func (e ExpiryDuration) ToDuration() time.Duration {
	switch e {
	case Expiry1Hour:
		return time.Hour
	case Expiry1Day:
		return 24 * time.Hour
	case Expiry1Month:
		return 30 * 24 * time.Hour
	case Expiry1Year:
		return 365 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}
