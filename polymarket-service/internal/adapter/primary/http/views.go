package http

import "github.com/JIeeiroSst/polymarket-service/internal/domain/model"

type marketView struct {
	model.Market
	YesPrice float64 `json:"yes_price"`
	NoPrice  float64 `json:"no_price"`
}

func toMarketView(m *model.Market) marketView {
	v := float64(m.ShareValue)
	return marketView{Market: *m, YesPrice: m.PriceOf(model.OutcomeYes) / v, NoPrice: m.PriceOf(model.OutcomeNo) / v}
}

type eventView struct {
	model.Event
	Markets []marketView `json:"markets"`
}

func toEventView(e *model.Event) eventView {
	markets := make([]marketView, len(e.Markets))
	for i := range e.Markets {
		markets[i] = toMarketView(&e.Markets[i])
	}
	return eventView{Event: *e, Markets: markets}
}

func toEventViews(events []model.Event) []eventView {
	out := make([]eventView, len(events))
	for i := range events {
		out[i] = toEventView(&events[i])
	}
	return out
}
