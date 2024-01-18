package services

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBService defines the operations performed on the database.
type DBService interface {
	FindOne(context.Context, interface{}, ...*options.FindOneOptions) *mongo.SingleResult
	Find(context.Context, interface{}, ...*options.FindOptions) (*mongo.Cursor, error)
	InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOneAndReplace(context.Context, interface{}, interface{}, ...*options.FindOneAndReplaceOptions) *mongo.SingleResult
	DeleteOne(context.Context, interface{}, ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}
