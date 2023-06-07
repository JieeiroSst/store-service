package card

type (
	CartItem struct {
		ProductId int
		Quantity  int
	}

	CartState struct {
		Items []CartItem
		Email string
	}

	UpdateCartMessage struct {
		Remove bool
		Item   CartItem
	}
)
