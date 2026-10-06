package domain

type Account struct {
	AccountID               string  `json:"account_id"`
	AccountEmail            string  `json:"account_email"`
	AccountStatus           string  `json:"account_status,omitempty"`
	PlanID                  string  `json:"plan_id"`
	PlanName                string  `json:"plan_name"`
	PlanMonthlyPrice        any     `json:"plan_monthly_price,omitempty"`
	PlanRenewalDate         *string `json:"plan_renewal_date"`
	SearchesPerMonth        int     `json:"searches_per_month"`
	PlanSearchesLeft        int     `json:"plan_searches_left"`
	ExtraCredits            int     `json:"extra_credits"`
	TotalSearchesLeft       int     `json:"total_searches_left"`
	ThisMonthUsage          int     `json:"this_month_usage"`
	ThisHourSearches        int     `json:"this_hour_searches"`
	LastHourSearches        int     `json:"last_hour_searches"`
	AccountRateLimitPerHour int     `json:"account_rate_limit_per_hour"`
}

type LocationQuery struct {
	Q     string
	Limit int
}

type Location struct {
	ID             string    `json:"id"`
	GoogleID       int64     `json:"google_id"`
	GoogleParentID int64     `json:"google_parent_id"`
	Name           string    `json:"name"`
	CanonicalName  string    `json:"canonical_name"`
	CountryCode    string    `json:"country_code"`
	TargetType     string    `json:"target_type"`
	Reach          int64     `json:"reach"`
	GPS            []float64 `json:"gps,omitempty"`
	Keys           []string  `json:"keys,omitempty"`
}

const MaxLocationLimit = 10

func (q *LocationQuery) Normalize() error {
	if q.Limit == 0 {
		q.Limit = 10
	}
	if q.Limit < 0 || q.Limit > MaxLocationLimit {
		return Invalid("limit must be between 1 and %d", MaxLocationLimit)
	}
	return nil
}
