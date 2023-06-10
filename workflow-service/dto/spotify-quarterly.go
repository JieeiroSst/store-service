package dto

type SpotifyQuarterly struct {
	Date                       string
	TotalRevenue               int
	CostOfRevenue              int
	GrossProfit                int
	PremiumRevenue             int
	PremiumCostRevenue         int
	PremiumGrossProfit         int
	AdRevenue                  int
	AdCostOfRevenue            int
	AdGrossProfit              int
	MAUs                       int
	PremiumMAUs                int
	AdMAUs                     int
	PremiumARPU                string
	SalesAndMarketingCost      int
	ResearchAndDevelopmentCost int
	GenrealAndAdminstraiveCost int
}