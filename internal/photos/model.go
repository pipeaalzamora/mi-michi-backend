package photos

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatPhoto struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	CatID     primitive.ObjectID `bson:"cat_id" json:"cat_id"`
	PhotoKey  string             `bson:"photo_key" json:"-"`
	PhotoURL  string             `bson:"-" json:"photo_url"`
	Caption   string             `bson:"caption" json:"caption"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
