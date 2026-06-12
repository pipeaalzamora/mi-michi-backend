package integrations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mi-michi/backend/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const cacheCollection = "external_cache"

type cacheRecord struct {
	Key       string    `bson:"key"`
	Payload   []byte    `bson:"payload"`
	ExpiresAt time.Time `bson:"expires_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func cacheJSON(ctx context.Context, key string, ttl time.Duration, target any, fetch func() (any, error)) error {
	if ttl > 0 {
		var record cacheRecord
		err := db.Col(cacheCollection).FindOne(ctx, bson.M{
			"key":        key,
			"expires_at": bson.M{"$gt": time.Now()},
		}).Decode(&record)
		if err == nil {
			if unmarshalErr := json.Unmarshal(record.Payload, target); unmarshalErr == nil {
				return nil
			}
		} else if err != mongo.ErrNoDocuments {
			return err
		}
	}

	value, err := fetch()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}

	if ttl <= 0 {
		return nil
	}

	now := time.Now()
	_, err = db.Col(cacheCollection).UpdateOne(ctx,
		bson.M{"key": key},
		bson.M{"$set": bson.M{
			"key":        key,
			"payload":    payload,
			"expires_at": now.Add(ttl),
			"updated_at": now,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}
