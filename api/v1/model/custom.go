package model

type Device struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	MaterialID uint   `json:"materialID"`
}

type Point struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	UpperLimit float64 `json:"upperLimit"`
	Nominal    float64 `json:"nominal"`
	LowerLimit float64 `json:"lowerLimit"`
	MaterialID uint    `json:"materialID"`
}
