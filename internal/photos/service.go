package photos

import (
	"context"
	"errors"
	"time"

	"github.com/mi-michi/backend/internal/db"
	"github.com/mi-michi/backend/internal/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection = "cat_photos"

func List(ctx context.Context, userID, catID string) ([]CatPhoto, error) {
	oid, err := primitive.ObjectIDFromHex(catID)
	if err != nil {
		return nil, errors.New("id de gato inválido")
	}
	cursor, err := db.Col(collection).Find(ctx,
		bson.M{"user_id": userID, "cat_id": oid},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		return nil, err
	}
	var photos []CatPhoto
	if err := cursor.All(ctx, &photos); err != nil {
		return nil, err
	}
	if photos == nil {
		photos = []CatPhoto{}
	}
	hydratePhotoURLs(ctx, photos)
	return photos, nil
}

func Create(ctx context.Context, userID, catID, photoKey, caption string) (*CatPhoto, error) {
	oid, err := primitive.ObjectIDFromHex(catID)
	if err != nil {
		return nil, errors.New("id de gato inválido")
	}
	now := time.Now()
	photo := CatPhoto{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		CatID:     oid,
		PhotoKey:  photoKey,
		Caption:   caption,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := db.Col(collection).InsertOne(ctx, photo); err != nil {
		return nil, err
	}
	hydratePhotoURL(ctx, &photo)
	return &photo, nil
}

func GetByID(ctx context.Context, userID, photoID string) (*CatPhoto, error) {
	oid, err := primitive.ObjectIDFromHex(photoID)
	if err != nil {
		return nil, errors.New("id de foto inválido")
	}
	var photo CatPhoto
	err = db.Col(collection).FindOne(ctx, bson.M{"_id": oid, "user_id": userID}).Decode(&photo)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	hydratePhotoURL(ctx, &photo)
	return &photo, nil
}

func Delete(ctx context.Context, userID, photoID string) error {
	photo, err := GetByID(ctx, userID, photoID)
	if err != nil {
		return err
	}
	if photo == nil {
		return errors.New("foto no encontrada")
	}
	if err := storage.DeleteCatPhoto(ctx, photo.PhotoKey); err != nil {
		return err
	}
	res, err := db.Col(collection).DeleteOne(ctx, bson.M{"_id": photo.ID, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("foto no encontrada")
	}
	return nil
}

func hydratePhotoURLs(ctx context.Context, photos []CatPhoto) {
	for i := range photos {
		hydratePhotoURL(ctx, &photos[i])
	}
}

func hydratePhotoURL(ctx context.Context, photo *CatPhoto) {
	if photo == nil || photo.PhotoKey == "" {
		return
	}
	url, err := storage.PresignCatPhoto(ctx, photo.PhotoKey, 6*time.Hour)
	if err != nil {
		return
	}
	photo.PhotoURL = url
}
