package vaccines

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

const collection = "vaccines"

func List(ctx context.Context, catID, userID string) ([]Vaccine, error) {
	cursor, err := db.Col(collection).Find(ctx,
		bson.M{"cat_id": catID, "user_id": userID},
		options.Find().SetSort(bson.M{"applied_date": -1}),
	)
	if err != nil {
		return nil, err
	}
	var vaccines []Vaccine
	if err = cursor.All(ctx, &vaccines); err != nil {
		return nil, err
	}
	if vaccines == nil {
		vaccines = []Vaccine{}
	}
	return vaccines, nil
}

// ListUpcoming devuelve vacunas con next_due_date en los próximos `days` días.
// Usado por el job de notificaciones.
func ListUpcoming(ctx context.Context, days int) ([]Vaccine, error) {
	now := time.Now()
	limit := now.AddDate(0, 0, days).Format("2006-01-02")
	today := now.Format("2006-01-02")

	cursor, err := db.Col(collection).Find(ctx, bson.M{
		"next_due_date": bson.M{
			"$gte": today,
			"$lte": limit,
		},
	})
	if err != nil {
		return nil, err
	}
	var vaccines []Vaccine
	if err = cursor.All(ctx, &vaccines); err != nil {
		return nil, err
	}
	return vaccines, nil
}

func Create(ctx context.Context, catID, userID string, req CreateVaccineRequest) (*Vaccine, error) {
	now := time.Now()
	v := Vaccine{
		ID:           primitive.NewObjectID(),
		CatID:        catID,
		UserID:       userID,
		Name:         req.Name,
		AppliedDate:  req.AppliedDate,
		NextDueDate:  req.NextDueDate,
		Veterinarian: req.Veterinarian,
		Notes:        req.Notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err := db.Col(collection).InsertOne(ctx, v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func Update(ctx context.Context, vacID, catID, userID string, req UpdateVaccineRequest) (*Vaccine, error) {
	oid, err := primitive.ObjectIDFromHex(vacID)
	if err != nil {
		return nil, errors.New("id inválido")
	}
	update := bson.M{"$set": bson.M{
		"name":          req.Name,
		"applied_date":  req.AppliedDate,
		"next_due_date": req.NextDueDate,
		"veterinarian":  req.Veterinarian,
		"notes":         req.Notes,
		"updated_at":    time.Now(),
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var v Vaccine
	err = db.Col(collection).FindOneAndUpdate(ctx,
		bson.M{"_id": oid, "cat_id": catID, "user_id": userID},
		update, opts,
	).Decode(&v)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &v, err
}

func Delete(ctx context.Context, vacID, catID, userID string) error {
	oid, err := primitive.ObjectIDFromHex(vacID)
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
		return errors.New("vacuna no encontrada")
	}
	return nil
}
