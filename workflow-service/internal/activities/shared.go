package activities

var SignalChannels = struct {
	ADD_TO_CART_CHANNEL      string
	REMOVE_FROM_CART_CHANNEL string
	UPDATE_EMAIL_CHANNEL     string
	CHECKOUT_CHANNEL         string
}{
	ADD_TO_CART_CHANNEL:      "ADD_TO_CART_CHANNEL",
	REMOVE_FROM_CART_CHANNEL: "REMOVE_FROM_CART_CHANNEL",
	UPDATE_EMAIL_CHANNEL:     "UPDATE_EMAIL_CHANNEL",
	CHECKOUT_CHANNEL:         "CHECKOUT_CHANNEL",
}

var RouteTypes = struct {
	ADD_TO_CART      string
	REMOVE_FROM_CART string
	UPDATE_EMAIL     string
	CHECKOUT         string
}{
	ADD_TO_CART:      "add_to_cart",
	REMOVE_FROM_CART: "remove_from_cart",
	UPDATE_EMAIL:     "update_email",
	CHECKOUT:         "checkout",
}

type RouteSignal struct {
	Route string
}

type AddToCartSignal struct {
	Route string
	Item  CartItem
}

type RemoveFromCartSignal struct {
	Route string
	Item  CartItem
}

type UpdateEmailSignal struct {
	Route string
	Email string
}

type CheckoutSignal struct {
	Route string
	Email string
}
