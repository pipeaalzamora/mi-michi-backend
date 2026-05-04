package vaccines

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vaccine struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CatID        string             `bson:"cat_id"        json:"cat_id"`
	UserID       string             `bson:"user_id"       json:"user_id"`
	Name         string             `bson:"name"          json:"name"`
	AppliedDate  string             `bson:"applied_date"  json:"applied_date"`
	NextDueDate  *string            `bson:"next_due_date" json:"next_due_date"`
	Veterinarian *string            `bson:"veterinarian"  json:"veterinarian"`
	Notes        *string            `bson:"notes"         json:"notes"`
	CreatedAt    time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"    json:"updated_at"`
}

type CreateVaccineRequest struct {
	Name         string  `json:"name"          binding:"required,min=1,max=200"`
	AppliedDate  string  `json:"applied_date"  binding:"required"`
	NextDueDate  *string `json:"next_due_date"`
	Veterinarian *string `json:"veterinarian"`
	Notes        *string `json:"notes"`
}

type UpdateVaccineRequest struct {
	Name         string  `json:"name"          binding:"required,min=1,max=200"`
	AppliedDate  string  `json:"applied_date"  binding:"required"`
	NextDueDate  *string `json:"next_due_date"`
	Veterinarian *string `json:"veterinarian"`
	Notes        *string `json:"notes"`
}
