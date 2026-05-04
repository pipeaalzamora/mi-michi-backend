package cats

import (
	"context"
	"errors"
	"time"

	"github.com/mi-michi/backend/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection = "cats"

func List(ctx context.Context, userID string) ([]Cat, error) {
	cursor, err := db.Col(collection).Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": 1}),
	)
	if err != nil {
		return nil, err
	}
	var cats []Cat
	if err = cursor.All(ctx, &cats); err != nil {
		return nil, err
	}
	if cats == nil {
		cats = []Cat{}
	}
	return cats, nil
}

func Create(ctx context.Context, userID string, req CreateCatRequest) (*Cat, error) {
	sex := req.Sex
	if sex == "" {
		sex = SexUnknown
	}
	now := time.Now()
	cat := Cat{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      req.Name,
		BirthDate: req.BirthDate,
		Breed:     req.Breed,
		Sex:       sex,
		Color:     req.Color,
		WeightKg:  req.WeightKg,
		Notes:     req.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := db.Col(collection).InsertOne(ctx, cat)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func GetByID(ctx context.Context, id, userID string) (*Cat, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("id inválido")
	}
	var cat Cat
	err = db.Col(collection).FindOne(ctx, bson.M{"_id": oid, "user_id": userID}).Decode(&cat)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &cat, err
}

func Update(ctx context.Context, id, userID string, req UpdateCatRequest) (*Cat, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("id inválido")
	}
	sex := req.Sex
	if sex == "" {
		sex = SexUnknown
	}
	update := bson.M{"$set": bson.M{
		"name":       req.Name,
		"birth_date": req.BirthDate,
		"breed":      req.Breed,
		"sex":        sex,
		"color":      req.Color,
		"weight_kg":  req.WeightKg,
		"notes":      req.Notes,
		"photo_url":  req.PhotoURL,
		"updated_at": time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var cat Cat
	err = db.Col(collection).FindOneAndUpdate(ctx, bson.M{"_id": oid, "user_id": userID}, update, opts).Decode(&cat)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &cat, err
}

func UpdatePhoto(ctx context.Context, id, userID, photoURL string) (*Cat, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("id inválido")
	}
	update := bson.M{"$set": bson.M{"photo_url": photoURL, "updated_at": time.Now()}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var cat Cat
	err = db.Col(collection).FindOneAndUpdate(ctx, bson.M{"_id": oid, "user_id": userID}, update, opts).Decode(&cat)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &cat, err
}

func Delete(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("id inválido")
	}
	res, err := db.Col(collection).DeleteOne(ctx, bson.M{"_id": oid, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("gato no encontrado")
	}
	return nil
}
