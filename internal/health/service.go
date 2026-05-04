package health

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

const collection = "health_logs"

func List(ctx context.Context, catID, userID string) ([]HealthLog, error) {
	cursor, err := db.Col(collection).Find(ctx,
		bson.M{"cat_id": catID, "user_id": userID},
		options.Find().SetSort(bson.M{"log_date": -1}).SetLimit(200),
	)
	if err != nil {
		return nil, err
	}
	var logs []HealthLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []HealthLog{}
	}
	return logs, nil
}

func Create(ctx context.Context, catID, userID string, req CreateLogRequest) (*HealthLog, error) {
	now := time.Now()
	log := HealthLog{
		ID:           primitive.NewObjectID(),
		CatID:        catID,
		UserID:       userID,
		LogType:      req.LogType,
		LogDate:      req.LogDate,
		Title:        req.Title,
		NumericValue: req.NumericValue,
		Description:  req.Description,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := db.Col(collection).InsertOne(ctx, log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func Update(ctx context.Context, logID, catID, userID string, req UpdateLogRequest) (*HealthLog, error) {
	oid, err := primitive.ObjectIDFromHex(logID)
	if err != nil {
		return nil, errors.New("id inválido")
	}
	update := bson.M{"$set": bson.M{
		"log_type":      req.LogType,
		"log_date":      req.LogDate,
		"title":         req.Title,
		"numeric_value": req.NumericValue,
		"description":   req.Description,
		"updated_at":    time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var log HealthLog
	err = db.Col(collection).FindOneAndUpdate(ctx,
		bson.M{"_id": oid, "cat_id": catID, "user_id": userID},
		update, opts,
	).Decode(&log)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &log, err
}

func Delete(ctx context.Context, logID, catID, userID string) error {
	oid, err := primitive.ObjectIDFromHex(logID)
	if err != nil {
		return errors.New("id inválido")
	}
	res, err := db.Col(collection).DeleteOne(ctx,
		bson.M{"_id": oid, "cat_id": catID, "user_id": userID},
	)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("registro no encontrado")
	}
	return nil
}
