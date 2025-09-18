package models

import (
	"time"
)

type Exercise struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	PrimaryMuscleGroup string    `json:"primary_muscle_group"`
	Equipment         string    `json:"equipment"`
	CreatedAt         time.Time `json:"created_at"`
}
