package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewGuard, NewCampaignService, NewGameService, NewCheckinService, NewFraudService, NewRewardRelay),
	fx.Invoke(registerRelay),
)
