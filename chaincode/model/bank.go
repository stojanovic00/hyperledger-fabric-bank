package model

type Bank struct {
	ID           string `json:"ID"`
	Name         string `json:"name"`
	Headquarters string `json:"headquarters"`
	Since        int    `json:"since"`
	PIB          int    `json:"pib"`
}
