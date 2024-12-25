package build

import (
	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
)

func BuildSubscription(req dto.Subscription) model.Subscription {
	if req.SubscriptionID == 0 {
		req.SubscriptionID = geared_id.GearedIntID()
	}
	return model.Subscription{
		SubscriptionID: req.SubscriptionID,
		Name:           req.Name,
		SubscribedFrom: req.SubscribedFrom,
		ValidUpto:      req.ValidUpto,
	}
}
