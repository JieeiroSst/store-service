package dynamodb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"

	appconfig "github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

type referralEventRepo struct {
	db    *dynamodb.Client
	table string
	log   *zap.Logger
}

func NewReferralEventRepo(db *dynamodb.Client, cfg *appconfig.Config, log *zap.Logger) ports.ReferralEventRepository {
	return &referralEventRepo{
		db:    db,
		table: cfg.DynamoDB.TableReferralEvents,
		log:   log.Named("event-repo"),
	}
}

func (r *referralEventRepo) Save(ctx context.Context, event *domain.ReferralEvent) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put event: %w", err)
	}
	r.log.Debug("event saved",
		zap.String("ref_code", event.RefCode),
		zap.String("event_type", string(event.EventType)),
	)
	return nil
}

func (r *referralEventRepo) FindByRefCode(ctx context.Context, refCode string) ([]*domain.ReferralEvent, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String("ref_code = :rc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rc": &types.AttributeValueMemberS{Value: refCode},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("query events for ref_code %s: %w", refCode, err)
	}
	return unmarshalEvents(out.Items)
}

func (r *referralEventRepo) FindByNewUserID(ctx context.Context, newUserID string) ([]*domain.ReferralEvent, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("new_user_id-index"),
		KeyConditionExpression: aws.String("new_user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: newUserID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query events for user %s: %w", newUserID, err)
	}
	return unmarshalEvents(out.Items)
}

func unmarshalEvents(items []map[string]types.AttributeValue) ([]*domain.ReferralEvent, error) {
	events := make([]*domain.ReferralEvent, 0, len(items))
	for _, item := range items {
		var e domain.ReferralEvent
		if err := attributevalue.UnmarshalMap(item, &e); err != nil {
			return nil, fmt.Errorf("unmarshal event: %w", err)
		}
		events = append(events, &e)
	}
	return events, nil
}

type rewardRepo struct {
	db    *dynamodb.Client
	table string
	log   *zap.Logger
}

func NewRewardRepo(db *dynamodb.Client, cfg *appconfig.Config, log *zap.Logger) ports.RewardRepository {
	return &rewardRepo{
		db:    db,
		table: cfg.DynamoDB.TableRewards,
		log:   log.Named("reward-repo"),
	}
}

func (r *rewardRepo) Save(ctx context.Context, reward *domain.ReferralReward) error {
	item, err := attributevalue.MarshalMap(reward)
	if err != nil {
		return fmt.Errorf("marshal reward: %w", err)
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put reward: %w", err)
	}
	r.log.Info("reward saved",
		zap.String("owner", reward.OwnerUserID),
		zap.String("ref_code", reward.RefCode),
		zap.Float64("value", reward.RewardValue),
	)
	return nil
}

func (r *rewardRepo) FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralReward, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String("owner_user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: ownerUserID},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("query rewards for owner %s: %w", ownerUserID, err)
	}

	rewards := make([]*domain.ReferralReward, 0, len(out.Items))
	for _, item := range out.Items {
		var rw domain.ReferralReward
		if err := attributevalue.UnmarshalMap(item, &rw); err != nil {
			return nil, fmt.Errorf("unmarshal reward: %w", err)
		}
		rewards = append(rewards, &rw)
	}
	return rewards, nil
}

type userStatsRepo struct {
	db    *dynamodb.Client
	table string
	log   *zap.Logger
}

func NewUserStatsRepo(db *dynamodb.Client, cfg *appconfig.Config, log *zap.Logger) ports.UserStatsRepository {
	return &userStatsRepo{
		db:    db,
		table: cfg.DynamoDB.TableUserStats,
		log:   log.Named("stats-repo"),
	}
}

func (r *userStatsRepo) Get(ctx context.Context, userID string) (*domain.UserReferralStats, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"sk":      &types.AttributeValueMemberS{Value: "STATS"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get stats for %s: %w", userID, err)
	}

	if out.Item == nil {
		return &domain.UserReferralStats{
			UserID: userID,
			SK:     "STATS",
		}, nil
	}

	var stats domain.UserReferralStats
	if err := attributevalue.UnmarshalMap(out.Item, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return &stats, nil
}

func (r *userStatsRepo) IncrementCounters(
	ctx context.Context,
	userID string,
	invited, installed, rewarded int64,
	rewardAmt float64,
) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"sk":      &types.AttributeValueMemberS{Value: "STATS"},
		},
		UpdateExpression: aws.String(
			"ADD total_invited :i, total_installed :inst, total_rewarded :r, total_reward_amt :amt " +
				"SET last_active_at = :ts",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":i":    &types.AttributeValueMemberN{Value: strconv.FormatInt(invited, 10)},
			":inst": &types.AttributeValueMemberN{Value: strconv.FormatInt(installed, 10)},
			":r":    &types.AttributeValueMemberN{Value: strconv.FormatInt(rewarded, 10)},
			":amt":  &types.AttributeValueMemberN{Value: strconv.FormatFloat(rewardAmt, 'f', 2, 64)},
			":ts":   &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		return fmt.Errorf("increment stats for %s: %w", userID, err)
	}
	return nil
}
