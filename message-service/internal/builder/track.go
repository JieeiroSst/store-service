package builder

import (
	"github.com/JIeeiroSst/message-service/model"
	"github.com/JIeeiroSst/message-service/pkg/snowflake"
)

func BuildTrackModel(data []byte, types, topic string) model.Track {
	return model.Track{
		ID:      snowflake.GearedID(),
		Topic:   topic,
		Type:    types,
		Message: string(data),
	}
}
