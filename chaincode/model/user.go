package model

type User struct {
	ID      string `json:"ID"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
}
