package consumer

import (
	platformkafka "github.com/JIeeiroSst/voucher-service/internal/platform/kafka"
	kafkago "github.com/segmentio/kafka-go"
)

type Reader = kafkago.Reader
type ReaderFactory = platformkafka.ReaderFactory
