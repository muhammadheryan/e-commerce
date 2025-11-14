package model

import "time"

// URL represents the url table entity
type URLEntity struct {
	ID          uint64     `db:"id" json:"id"`
	UserID      uint64     `db:"user_id" json:"user_id"`
	ShortURL    string     `db:"short_url" json:"short_url"`
	OriginalURL string     `db:"original_url" json:"original_url"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type URLFilter struct {
	ID       uint64
	ShortURL string
}

type GetURLResponse struct {
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CreateURLShortnerRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}
