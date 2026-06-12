package profile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa un usuario en MongoDB.
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"    json:"id"`
	FirebaseUID string             `bson:"firebase_uid,omitempty" json:"firebase_uid,omitempty"`
	GoogleID    string             `bson:"google_id"        json:"google_id"`
	Email       string             `bson:"email"            json:"email"`
	DisplayName string             `bson:"display_name"     json:"display_name"`
	Picture     string             `bson:"picture"          json:"picture"`
	CreatedAt   time.Time          `bson:"created_at"       json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"       json:"updated_at"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=1,max=80"`
}
