package model

type BasketLineAttribute struct {
	ID       int    `db:"id"`
	BasketID int    `db:"basket_id"`
	LineID   int    `db:"line_id"`
	OptionID int    `db:"option_id"`
	Value    string `db:"value"`
}
