package models

type Currency struct {
	ID       int64  `json:"id"`
	Slug     string `json:"slug"`
	IsActive bool   `json:"-"`
}
