package models

import (
	"time"
)

type UserSession struct {
	ID        string    `db:"id" json:"id"`                 // SHA256 hex string
	UserID    int       `db:"user_id" json:"user_id"`       // user id
	Token     string    `db:"token" json:"token"`           // token string
	IP        string    `db:"ip" json:"ip"`                 // IP address
	UserAgent string    `db:"user_agent" json:"user_agent"` // user agent string
	Expires   time.Time `db:"expires" json:"expires"`       // expiry time
	LoginAt   time.Time `db:"login_at" json:"login_at"`     // login time
	IsCurrent bool      `db:"is_current" json:"is_current"` // is current session
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
