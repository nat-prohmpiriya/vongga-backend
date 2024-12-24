package repository

import (
	"context"

	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/domain"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type subPostRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewSubPostRepository(db *mongo.Database) domain.SubPostRepository {
	return &subPostRepository{
		db:         db,
		collection: db.Collection("subposts"),
	}
}

func (r *subPostRepository) Create(subPost *domain.SubPost) error {
	logger := utils.NewLogger("SubPostRepository.Create")
	logger.LogInput(subPost)

	_, err := r.collection.InsertOne(context.Background(), subPost)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("SubPost created successfully", nil)
	return nil
}

func (r *subPostRepository) Update(subPost *domain.SubPost) error {
	logger := utils.NewLogger("SubPostRepository.Update")
	logger.LogInput(subPost)

	filter := bson.M{"_id": subPost.ID}
	update := bson.M{"$set": subPost}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("SubPost updated successfully", nil)
	return nil
}

func (r *subPostRepository) Delete(id primitive.ObjectID) error {
	logger := utils.NewLogger("SubPostRepository.Delete")
	logger.LogInput(id)

	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput("SubPost deleted successfully", nil)
	return nil
}

func (r *subPostRepository) FindByID(id primitive.ObjectID) (*domain.SubPost, error) {
	logger := utils.NewLogger("SubPostRepository.FindByID")
	logger.LogInput(id)

	var subPost domain.SubPost
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(context.Background(), filter).Decode(&subPost)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPost, nil)
	return &subPost, nil
}

func (r *subPostRepository) FindByParentID(parentID primitive.ObjectID, limit, offset int) ([]domain.SubPost, error) {
	logger := utils.NewLogger("SubPostRepository.FindByParentID")
	input := map[string]interface{}{
		"parentID": parentID,
		"limit":    limit,
		"offset":   offset,
	}
	logger.LogInput(input)

	var subPosts []domain.SubPost
	filter := bson.M{"parentId": parentID}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}
	findOptions.SetSort(bson.D{{Key: "order", Value: 1}, {Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &subPosts)
	if err != nil {
		logger.LogOutput(nil, err)
		return nil, err
	}

	logger.LogOutput(subPosts, nil)
	return subPosts, nil
}

func (r *subPostRepository) UpdateOrder(parentID primitive.ObjectID, orders map[primitive.ObjectID]int) error {
	logger := utils.NewLogger("SubPostRepository.UpdateOrder")
	input := map[string]interface{}{
		"parentID": parentID,
		"orders":   orders,
	}
	logger.LogInput(input)

	// Use bulk write to update multiple documents efficiently
	var operations []mongo.WriteModel
	for subPostID, order := range orders {
		filter := bson.M{"_id": subPostID, "parentId": parentID}
		update := bson.M{"$set": bson.M{"order": order}}
		operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		operations = append(operations, operation)
	}

	if len(operations) > 0 {
		result, err := r.collection.BulkWrite(context.Background(), operations)
		if err != nil {
			logger.LogOutput(nil, err)
			return err
		}
		logger.LogOutput(map[string]interface{}{
			"message":           "SubPosts order updated successfully",
			"matchedCount":     result.MatchedCount,
			"modifiedCount":    result.ModifiedCount,
		}, nil)
		return nil
	}

	logger.LogOutput("No orders to update", nil)
	return nil
}

func (r *subPostRepository) DeleteByParentID(parentID primitive.ObjectID) error {
	logger := utils.NewLogger("SubPostRepository.DeleteByParentID")
	logger.LogInput(parentID)

	filter := bson.M{"parentId": parentID}
	result, err := r.collection.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.LogOutput(nil, err)
		return err
	}

	logger.LogOutput(map[string]interface{}{
		"message":       "SubPosts deleted successfully",
		"deletedCount": result.DeletedCount,
	}, nil)
	return nil
}
