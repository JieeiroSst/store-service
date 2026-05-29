package dynamodb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"

	appconfig "github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

type referralLinkRepo struct {
	db    *dynamodb.Client
	table string
	log   *zap.Logger
}

func NewReferralLinkRepo(db *dynamodb.Client, cfg *appconfig.Config, log *zap.Logger) ports.ReferralLinkRepository {
	return &referralLinkRepo{
		db:    db,
		table: cfg.DynamoDB.TableReferralLinks,
		log:   log.Named("link-repo"),
	}
}

func (r *referralLinkRepo) Save(ctx context.Context, link *domain.ReferralLink) error {
	item, err := attributevalue.MarshalMap(link)
	if err != nil {
		return fmt.Errorf("marshal referral link: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(ref_code)"),
	})
	if err != nil {
		return fmt.Errorf("put referral link: %w", err)
	}

	r.log.Debug("referral link saved", zap.String("ref_code", link.RefCode))
	return nil
}

func (r *referralLinkRepo) FindByRefCode(ctx context.Context, refCode string) (*domain.ReferralLink, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String("ref_code = :rc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rc": &types.AttributeValueMemberS{Value: refCode},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query ref_code %s: %w", refCode, err)
	}

	if len(out.Items) == 0 {
		return nil, fmt.Errorf("ref_code %s: %w", refCode, domain.ErrNotFound)
	}

	var link domain.ReferralLink
	if err := attributevalue.UnmarshalMap(out.Items[0], &link); err != nil {
		return nil, fmt.Errorf("unmarshal referral link: %w", err)
	}

	return &link, nil
}

func (r *referralLinkRepo) FindByOwnerUserID(
	ctx context.Context,
	ownerUserID string,
	limit int,
	cursor string,
) ([]*domain.ReferralLink, string, error) {

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("owner_user_id-index"),
		KeyConditionExpression: aws.String("owner_user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: ownerUserID},
		},
		Limit:            aws.Int32(int32(limit)),
		ScanIndexForward: aws.Bool(false), // newest first
	}

	// Decode cursor (base64 JSON of LastEvaluatedKey)
	if cursor != "" {
		var startKey map[string]types.AttributeValue
		if err := decodeCursor(cursor, &startKey); err == nil {
			input.ExclusiveStartKey = startKey
		}
	}

	out, err := r.db.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("query links by owner %s: %w", ownerUserID, err)
	}

	links := make([]*domain.ReferralLink, 0, len(out.Items))
	for _, item := range out.Items {
		var link domain.ReferralLink
		if err := attributevalue.UnmarshalMap(item, &link); err != nil {
			r.log.Warn("unmarshal link failed", zap.Error(err))
			continue
		}
		links = append(links, &link)
	}

	nextCursor := ""
	if out.LastEvaluatedKey != nil {
		nextCursor, _ = encodeCursor(out.LastEvaluatedKey)
	}

	return links, nextCursor, nil
}

func (r *referralLinkRepo) UpdateStatus(
	ctx context.Context,
	refCode string,
	createdAt string,
	status domain.ReferralStatus,
) error {
	_, err := r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"ref_code":   &types.AttributeValueMemberS{Value: refCode},
			"created_at": &types.AttributeValueMemberS{Value: createdAt},
		},
		UpdateExpression:    aws.String("SET #s = :status"),
		ConditionExpression: aws.String("#s = :active"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
			":active": &types.AttributeValueMemberS{Value: string(domain.StatusActive)},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return fmt.Errorf("update status for ref_code %s: %w", refCode, domain.ErrLinkNotActive)
		}
		return fmt.Errorf("update status for ref_code %s: %w", refCode, err)
	}
	return nil
}

func encodeCursor(key map[string]types.AttributeValue) (string, error) {
	simple := make(map[string]string)
	for k, v := range key {
		if sv, ok := v.(*types.AttributeValueMemberS); ok {
			simple[k] = sv.Value
		}
	}
	b, err := json.Marshal(simple)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func decodeCursor(cursor string, out *map[string]types.AttributeValue) error {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return err
	}
	var simple map[string]string
	if err := json.Unmarshal(decoded, &simple); err != nil {
		return err
	}
	m := make(map[string]types.AttributeValue)
	for k, v := range simple {
		m[k] = &types.AttributeValueMemberS{Value: v}
	}
	*out = m
	return nil
}
