package domain

import "time"

type Review struct {
	ID         int64
	HomestayID int64
	UserID     int64
	BookingID  int64 
	LeaseID    int64
	Rating     int
	Comment    string
	CreatedAt  time.Time
}

const (
	ratingPrior  = 3.5 // what a homestay with no reviews is assumed to be worth
	ratingWeight = 5.0 // how many reviews it takes to outweigh the prior
)

func BayesianRating(avg float64, count int) float64 {
	n := float64(count)
	return (n*avg + ratingWeight*ratingPrior) / (n + ratingWeight)
}

// 1 bronze .. 4 platinum
func TierBoost(tier int) float64 {
	switch {
	case tier >= 4:
		return 0.5
	case tier == 3:
		return 0.3
	case tier == 2:
		return 0.15
	}
	return 0
}

func RankScore(avg float64, count, hostTier int) float64 {
	return BayesianRating(avg, count) + TierBoost(hostTier)
}

func HostPoints(rating int) int64 {
	switch rating {
	case 5:
		return 200
	case 4:
		return 100
	}
	return 0
}
