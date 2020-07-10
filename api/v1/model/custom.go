package model

type Device struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	MaterialID uint   `json:"materialID"`
}
