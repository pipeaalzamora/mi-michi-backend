package cats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatSex string

const (
	SexMale    CatSex = "macho"
	SexFemale  CatSex = "hembra"
	SexUnknown CatSex = "desconocido"
)

type Cat struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id"       json:"user_id"`
	Name      string             `bson:"name"          json:"name"`
	BirthDate *string            `bson:"birth_date"    json:"birth_date"`
	Breed     *string            `bson:"breed"         json:"breed"`
	Sex       CatSex             `bson:"sex"           json:"sex"`
	Color     *string            `bson:"color"         json:"color"`
	WeightKg  *float64           `bson:"weight_kg"     json:"weight_kg"`
	PhotoURL  *string            `bson:"photo_url"     json:"photo_url"`
	PhotoKey  *string            `bson:"photo_key"     json:"-"`
	Notes     *string            `bson:"notes"         json:"notes"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}

type CreateCatRequest struct {
	Name      string   `json:"name"       binding:"required,min=1,max=100"`
	BirthDate *string  `json:"birth_date"`
	Breed     *string  `json:"breed"`
	Sex       CatSex   `json:"sex"`
	Color     *string  `json:"color"`
	WeightKg  *float64 `json:"weight_kg"`
	Notes     *string  `json:"notes"`
}

type UpdateCatRequest struct {
	Name      string   `json:"name"       binding:"required,min=1,max=100"`
	BirthDate *string  `json:"birth_date"`
	Breed     *string  `json:"breed"`
	Sex       CatSex   `json:"sex"`
	Color     *string  `json:"color"`
	WeightKg  *float64 `json:"weight_kg"`
	Notes     *string  `json:"notes"`
	PhotoURL  *string  `json:"photo_url"`
}
