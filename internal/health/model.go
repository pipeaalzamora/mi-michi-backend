package health

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogType string

const (
	LogPeso      LogType = "peso"
	LogVisitaVet LogType = "visita_vet"
	LogSintoma   LogType = "sintoma"
	LogMedicamento LogType = "medicamento"
	LogOtro      LogType = "otro"
)

type HealthLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CatID        string             `bson:"cat_id"        json:"cat_id"`
	UserID       string             `bson:"user_id"       json:"user_id"`
	LogType      LogType            `bson:"log_type"      json:"log_type"`
	LogDate      string             `bson:"log_date"      json:"log_date"`
	NumericValue *float64           `bson:"numeric_value" json:"numeric_value"`
	Title        string             `bson:"title"         json:"title"`
	Description  *string            `bson:"description"   json:"description"`
	CreatedAt    time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"    json:"updated_at"`
}

type CreateLogRequest struct {
	LogType      LogType  `json:"log_type"      binding:"required"`
	LogDate      string   `json:"log_date"      binding:"required"`
	Title        string   `json:"title"         binding:"required,min=1,max=200"`
	NumericValue *float64 `json:"numeric_value"`
	Description  *string  `json:"description"`
}

type UpdateLogRequest struct {
	LogType      LogType  `json:"log_type"      binding:"required"`
	LogDate      string   `json:"log_date"      binding:"required"`
	Title        string   `json:"title"         binding:"required,min=1,max=200"`
	NumericValue *float64 `json:"numeric_value"`
	Description  *string  `json:"description"`
}
