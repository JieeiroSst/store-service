package dto

type Game struct {
	ID            string
	Rated         string
	CreatedAt     int
	LastMoveAt    int
	Turns         int
	VictoryStatus string
	Winner        string
	IncrementCode string
	WhiteId       string
	WhiteRating   int
	BlackId       string
	BlackRating   int
	Moves         string
	OpeningEco    string
	OpeningName   string
	OpeningPly    int
}
