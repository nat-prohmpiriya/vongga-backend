package repository

import (
	"context"
	"time"

	"vongga_api/internal/domain"
	"vongga_api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.opentelemetry.io/otel/trace"
)

type reactionRepository struct {
	db     *mongo.Database
	tracer trace.Tracer
}

func NewReactionRepository(db *mongo.Database, tracer trace.Tracer) domain.ReactionRepository {
	return &reactionRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *reactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.Create")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(reaction)

	reaction.CreatedAt = time.Now()
	reaction.UpdatedAt = time.Now()

	result, err := r.db.Collection("reactions").InsertOne(ctx, reaction)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	reaction.ID = result.InsertedID.(primitive.ObjectID)
	logger.Output(reaction, nil)
	return nil
}

func (r *reactionRepository) Update(ctx context.Context, reaction *domain.Reaction) error {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.Update")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(reaction)

	reaction.UpdatedAt = time.Now()

	filter := bson.M{"_id": reaction.ID}
	update := bson.M{"$set": reaction}

	_, err := r.db.Collection("reactions").UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	logger.Output(reaction, nil)
	return nil
}

func (r *reactionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.Delete")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"isActive":  false,
		},
	}

	_, err := r.db.Collection("reactions").UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Output("1", err)
		return err
	}

	logger.Output(map[string]interface{}{"id": id}, nil)
	return nil
}

func (r *reactionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.FindByID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(id)

	var reaction domain.Reaction
	err := r.db.Collection("reactions").FindOne(ctx, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&reaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output("1", domain.ErrNotFound)
			return nil, domain.ErrNotFound
		}
		logger.Output("2", err)
		return nil, err
	}

	logger.Output(&reaction, nil)
	return &reaction, nil
}

func (r *reactionRepository) FindByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.FindByPostID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"postID": postID,
		"limit":  limit,
		"offset": offset,
	})

	opts := options.Find().
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(ctx, bson.M{"postId": postID, "deletedAt": bson.M{"$exists": false}}, opts)
	if err != nil {
		logger.Output("1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []domain.Reaction
	if err = cursor.All(ctx, &reactions); err != nil {
		logger.Output("2", err)
		return nil, err
	}

	logger.Output(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByCommentID(ctx context.Context, commentID primitive.ObjectID, limit, offset int) ([]domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.FindByCommentID")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"commentID": commentID,
		"limit":     limit,
		"offset":    offset,
	})

	opts := options.Find().
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.db.Collection("reactions").Find(ctx, bson.M{"commentId": commentID, "deletedAt": bson.M{"$exists": false}}, opts)
	if err != nil {
		logger.Output("1", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []domain.Reaction
	if err = cursor.All(ctx, &reactions); err != nil {
		logger.Output("2", err)
		return nil, err
	}

	logger.Output(reactions, nil)
	return reactions, nil
}

func (r *reactionRepository) FindByUserAndTarget(ctx context.Context, userID, postID primitive.ObjectID, commentID *primitive.ObjectID) (*domain.Reaction, error) {
	ctx, span := r.tracer.Start(ctx, "ReactionRepository.FindByUserAndTarget")
	defer span.End()
	logger := utils.NewTraceLogger(span)
	logger.Input(map[string]interface{}{
		"userID":    userID,
		"postID":    postID,
		"commentID": commentID,
	})

	filter := bson.M{
		"userId": userID,
		"postId": postID,
	}
	if commentID != nil {
		filter["commentId"] = commentID
	}

	var reaction domain.Reaction
	err := r.db.Collection("reactions").FindOne(ctx, filter).Decode(&reaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Output(nil, nil)
			return nil, nil
		}
		logger.Output("1", err)
		return nil, err
	}

	logger.Output(&reaction, nil)
	return &reaction, nil
}
