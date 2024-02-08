package model

type Currency int8

const (
	EUR Currency = iota
	RSD
)

type BankAccount struct {
	ID       string   `json:"ID"`
	Balance  float64  `json:"balance"`
	Currency Currency `json:"currency"`
	Cards    []string `json:"cards"`

	Bank   Bank   `json:"bank"`
	UserID string `json:"user_id"`
}
