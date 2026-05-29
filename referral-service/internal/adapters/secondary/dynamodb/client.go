package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/fx"
	"go.uber.org/zap"

	appconfig "github.com/referral/service/internal/config"
)

var Module = fx.Options(
	fx.Provide(
		NewClient,
		NewReferralLinkRepo,
		NewReferralEventRepo,
		NewRewardRepo,
		NewUserStatsRepo,
	),
)

func NewClient(cfg *appconfig.Config, log *zap.Logger) (*dynamodb.Client, error) {
	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.AWS.Region),
	}

	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWS.AccessKeyID,
				cfg.AWS.SecretAccessKey,
				"",
			),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, err
	}

	clientOpts := []func(*dynamodb.Options){}

	if cfg.AWS.DynamoDBEndpoint != "" {
		log.Info("using local DynamoDB endpoint", zap.String("endpoint", cfg.AWS.DynamoDBEndpoint))
		clientOpts = append(clientOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.AWS.DynamoDBEndpoint)
		})
	}

	client := dynamodb.NewFromConfig(awsCfg, clientOpts...)

	log.Info("DynamoDB client ready",
		zap.String("region", cfg.AWS.Region),
		zap.Bool("local", cfg.AWS.DynamoDBEndpoint != ""),
	)

	return client, nil
}
