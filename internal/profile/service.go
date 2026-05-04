package profile

import (
	"context"
	"time"

	"github.com/mi-michi/backend/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection = "users"

// UpsertByGoogle crea o actualiza el usuario por su Google ID.
// Devuelve el usuario resultante.
func UpsertByGoogle(ctx context.Context, googleID, email, name, picture string) (*User, error) {
	col := db.Col(collection)
	now := time.Now()

	filter := bson.M{"google_id": googleID}
	update := bson.M{
		"$set": bson.M{
			"email":      email,
			"picture":    picture,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"google_id":    googleID,
			"display_name": name,
			"created_at":   now,
		},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var user User
	err := col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID devuelve un usuario por su ObjectID.
func GetByID(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var user User
	err = db.Col(collection).FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// UpdateDisplayName actualiza el nombre visible del usuario.
func UpdateDisplayName(ctx context.Context, id, displayName string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{"$set": bson.M{"display_name": displayName, "updated_at": time.Now()}}
	var user User
	err = db.Col(collection).FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&user)
	return &user, err
}
