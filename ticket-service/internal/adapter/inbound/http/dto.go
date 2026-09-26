package http

import (
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

// ---- events

type eventRequest struct {
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Category          string    `json:"category"`
	City              string    `json:"city"`
	Venue             string    `json:"venue"`
	Address           string    `json:"address"`
	BannerURL         string    `json:"banner_url"`
	StartsAt          time.Time `json:"starts_at"`
	EndsAt            time.Time `json:"ends_at"`
	Currency          string    `json:"currency"`
	WalletID          string    `json:"wallet_id"`
	RefundCutoffHours int       `json:"refund_cutoff_hours"`
	Transferable      *bool     `json:"transferable"`       // default true
	ResaleCapPercent  int       `json:"resale_cap_percent"` // 0: holders may not resell; 100: up to face value
}

func (r eventRequest) input() inbound.EventInput {
	return inbound.EventInput{Title: r.Title, Description: r.Description, Category: r.Category, City: r.City, Venue: r.Venue,
		Address: r.Address, BannerURL: r.BannerURL, StartsAt: r.StartsAt, EndsAt: r.EndsAt, Currency: r.Currency,
		WalletID: r.WalletID, RefundCutoffHours: r.RefundCutoffHours, Transferable: r.Transferable == nil || *r.Transferable, ResaleCapPercent: r.ResaleCapPercent}
}

type eventResponse struct {
	ID                int64                `json:"id"`
	OrganizerID       int64                `json:"organizer_id"`
	Title             string               `json:"title"`
	Description       string               `json:"description"`
	Category          string               `json:"category"`
	City              string               `json:"city"`
	Venue             string               `json:"venue"`
	Address           string               `json:"address,omitempty"`
	BannerURL         string               `json:"banner_url,omitempty"`
	StartsAt          time.Time            `json:"starts_at"`
	EndsAt            time.Time            `json:"ends_at"`
	Currency          string               `json:"currency"`
	Status            string               `json:"status"`
	ReviewNote        string               `json:"review_note,omitempty"`
	Featured          bool                 `json:"featured"`
	WalletID          string               `json:"wallet_id,omitempty"`
	RefundCutoffHours int                  `json:"refund_cutoff_hours"`
	Transferable      bool                 `json:"transferable"`
	ResaleCapPercent  int                  `json:"resale_cap_percent"`
	SeriesID          int64                `json:"series_id,omitempty"`
	VenueID           int64                `json:"venue_id,omitempty"`
	MinPrice          *int64               `json:"min_price,omitempty"`
	SoldOut           bool                 `json:"sold_out"`
	NextSessionAt     *time.Time           `json:"next_session_at,omitempty"`
	Sessions          []sessionResponse    `json:"sessions,omitempty"`
	TicketTypes       []ticketTypeResponse `json:"ticket_types,omitempty"`
}

func toEventResponse(e domain.Event) eventResponse {
	now := time.Now()
	out := eventResponse{ID: e.ID, OrganizerID: e.OrganizerID, Title: e.Title, Description: e.Description, Category: e.Category,
		City: e.City, Venue: e.Venue, Address: e.Address, BannerURL: e.BannerURL, StartsAt: e.StartsAt, EndsAt: e.EndsAt,
		Currency: e.Currency, Status: e.Status.Name(), ReviewNote: e.ReviewNote, Featured: e.Featured, WalletID: e.WalletID,
		RefundCutoffHours: e.RefundCutoffHours, Transferable: e.Transferable, ResaleCapPercent: e.ResaleCapPercent, SeriesID: e.SeriesID, VenueID: e.VenueID, MinPrice: e.MinPrice, SoldOut: e.SoldOut}
	out.NextSessionAt = e.NextSession
	for _, x := range e.Sessions {
		out.Sessions = append(out.Sessions, toSessionResponse(x))
	}
	for _, t := range e.TicketTypes {
		out.TicketTypes = append(out.TicketTypes, toTypeResponse(t, now))
	}
	return out
}

// ---- ticket types, seats, promotions

type ticketTypeRequest struct {
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Price        int64      `json:"price"`
	Total        int        `json:"total"`
	MinPerOrder  int        `json:"min_per_order"`
	MaxPerOrder  int        `json:"max_per_order"`
	MaxPerUser   int        `json:"max_per_user"`
	SaleStartsAt *time.Time `json:"sale_starts_at"`
	SaleEndsAt   *time.Time `json:"sale_ends_at"`
	Seated       bool       `json:"seated"`
	Active       *bool      `json:"active"`
	SortOrder    int        `json:"sort_order"`
	SessionID    int64      `json:"session_id"` // the showtime it sells; default the first one still to come
}

func (r ticketTypeRequest) input() inbound.TicketTypeInput {
	return inbound.TicketTypeInput{Name: r.Name, Description: r.Description, Price: r.Price, Total: r.Total,
		MinPerOrder: r.MinPerOrder, MaxPerOrder: r.MaxPerOrder, MaxPerUser: r.MaxPerUser, SaleStartsAt: r.SaleStartsAt,
		SaleEndsAt: r.SaleEndsAt, Seated: r.Seated, Active: r.Active == nil || *r.Active, SortOrder: r.SortOrder, SessionID: r.SessionID}
}

type ticketTypeResponse struct {
	ID           int64      `json:"id"`
	EventID      int64      `json:"event_id"`
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	Price        int64      `json:"price"`
	Total        int        `json:"total"`
	Available    int        `json:"available"`
	Sold         int        `json:"sold"`
	Held         int        `json:"held"`
	MinPerOrder  int        `json:"min_per_order"`
	MaxPerOrder  int        `json:"max_per_order"`
	MaxPerUser   int        `json:"max_per_user"`
	SaleStartsAt *time.Time `json:"sale_starts_at,omitempty"`
	SaleEndsAt   *time.Time `json:"sale_ends_at,omitempty"`
	Seated       bool       `json:"seated"`
	Active       bool       `json:"active"`
	SortOrder    int        `json:"sort_order"`
	SessionID    int64      `json:"session_id,omitempty"`
	OnSale       bool       `json:"on_sale"`
	SoldOut      bool       `json:"sold_out"`
}

func toTypeResponse(t domain.TicketType, now time.Time) ticketTypeResponse {
	return ticketTypeResponse{ID: t.ID, EventID: t.EventID, Name: t.Name, Description: t.Description, Price: t.Price,
		Total: t.Total, Available: t.Available, Sold: t.Sold, Held: t.Held(), MinPerOrder: t.MinPerOrder, MaxPerOrder: t.MaxPerOrder,
		MaxPerUser: t.MaxPerUser, SaleStartsAt: t.SaleStartsAt, SaleEndsAt: t.SaleEndsAt, Seated: t.Seated, Active: t.Active,
		SortOrder: t.SortOrder, SessionID: t.SessionID, OnSale: t.OnSale(now) && t.Available > 0, SoldOut: t.Available == 0}
}

type seatLayoutRequest struct {
	Section     string  `json:"section"`
	Rows        int     `json:"rows"`
	SeatsPerRow int     `json:"seats_per_row"`
	FirstRow    string  `json:"first_row"`
	OffsetX     float64 `json:"offset_x"`
	OffsetY     float64 `json:"offset_y"`
}

type seatResponse struct {
	ID           int64    `json:"id"`
	TicketTypeID int64    `json:"ticket_type_id,omitempty"`
	Accessible   bool     `json:"accessible,omitempty"`
	Section      string   `json:"section,omitempty"`
	Row          string   `json:"row"`
	Number       int      `json:"number"`
	Label        string   `json:"label"`
	Status       string   `json:"status"`
	X            *float64 `json:"x,omitempty"`
	Y            *float64 `json:"y,omitempty"`
}

func toSeatResponse(s domain.Seat) seatResponse {
	return seatResponse{ID: s.ID, TicketTypeID: s.TicketTypeID, Accessible: s.Accessible, Section: s.Section, Row: s.Row, Number: s.Number, Label: s.Label(), Status: s.Status.Name(), X: s.X, Y: s.Y}
}

type promotionRequest struct {
	Code       string     `json:"code"`
	Kind       string     `json:"kind"` // "percent" or "fixed"
	Value      int64      `json:"value"`
	MaxUses    int        `json:"max_uses"`
	MinTickets int        `json:"min_tickets"`
	ValidFrom  *time.Time `json:"valid_from"`
	ValidTo    *time.Time `json:"valid_to"`
	Active     *bool      `json:"active"`
}

func (r promotionRequest) input() inbound.PromotionInput {
	var kind domain.PromoKind
	switch r.Kind {
	case "percent":
		kind = domain.PromoPercent
	case "fixed":
		kind = domain.PromoFixed
	}
	return inbound.PromotionInput{Code: r.Code, Kind: kind, Value: r.Value, MaxUses: r.MaxUses, MinTickets: r.MinTickets,
		ValidFrom: r.ValidFrom, ValidTo: r.ValidTo, Active: r.Active == nil || *r.Active}
}

type promotionResponse struct {
	ID         int64      `json:"id"`
	Code       string     `json:"code"`
	Kind       string     `json:"kind"`
	Value      int64      `json:"value"`
	MaxUses    int        `json:"max_uses"`
	Used       int        `json:"used"`
	MinTickets int        `json:"min_tickets"`
	ValidFrom  *time.Time `json:"valid_from,omitempty"`
	ValidTo    *time.Time `json:"valid_to,omitempty"`
	Active     bool       `json:"active"`
}

func toPromotionResponse(p domain.Promotion) promotionResponse {
	kind := "percent"
	if p.Kind == domain.PromoFixed {
		kind = "fixed"
	}
	return promotionResponse{ID: p.ID, Code: p.Code, Kind: kind, Value: p.Value, MaxUses: p.MaxUses, Used: p.Used,
		MinTickets: p.MinTickets, ValidFrom: p.ValidFrom, ValidTo: p.ValidTo, Active: p.Active}
}

// ---- orders and tickets

type reserveItemRequest struct {
	TicketTypeID int64   `json:"ticket_type_id"`
	Quantity     int     `json:"quantity"`
	SeatIDs      []int64 `json:"seat_ids"`
}

type reserveRequest struct {
	EventID    int64                `json:"event_id"`
	Items      []reserveItemRequest `json:"items"`
	PromoCode  string               `json:"promo_code"`
	BuyerName  string               `json:"buyer_name"`
	BuyerEmail string               `json:"buyer_email"`
	BuyerPhone string               `json:"buyer_phone"`
	RequestID  string               `json:"request_id"`
}

func (r reserveRequest) command() inbound.ReserveCommand {
	items := make([]inbound.ReserveItem, len(r.Items))
	for i, it := range r.Items {
		items[i] = inbound.ReserveItem{TicketTypeID: it.TicketTypeID, Quantity: it.Quantity, SeatIDs: it.SeatIDs}
	}
	return inbound.ReserveCommand{EventID: r.EventID, Items: items, PromoCode: r.PromoCode, BuyerName: r.BuyerName,
		BuyerEmail: r.BuyerEmail, BuyerPhone: r.BuyerPhone, RequestID: r.RequestID}
}

type payRequest struct {
	Method   string `json:"method"`   // "wallet" or "gateway"
	Provider string `json:"provider"` // gateway only
}

type orderItemResponse struct {
	TicketTypeID int64   `json:"ticket_type_id"`
	Name         string  `json:"name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    int64   `json:"unit_price"`
	SeatIDs      []int64 `json:"seat_ids,omitempty"`
}

type orderResponse struct {
	ID            int64               `json:"id"`
	EventID       int64               `json:"event_id"`
	UserID        int64               `json:"user_id"`
	Status        string              `json:"status"`
	Currency      string              `json:"currency"`
	Subtotal      int64               `json:"subtotal"`
	Discount      int64               `json:"discount"`
	Total         int64               `json:"total"`
	PromoCode     string              `json:"promo_code,omitempty"`
	BuyerName     string              `json:"buyer_name"`
	BuyerEmail    string              `json:"buyer_email"`
	BuyerPhone    string              `json:"buyer_phone,omitempty"`
	SessionID     int64               `json:"session_id,omitempty"`
	ExpiresAt     time.Time           `json:"expires_at"`
	PaymentMethod string              `json:"payment_method,omitempty"`
	PaidAt        *time.Time          `json:"paid_at,omitempty"`
	RefundAmount  int64               `json:"refund_amount,omitempty"`
	StatusNote    string              `json:"status_note,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	Items         []orderItemResponse `json:"items"`
	Tickets       []ticketResponse    `json:"tickets,omitempty"`
}

func toOrderResponse(x domain.Order) orderResponse {
	out := orderResponse{ID: x.ID, EventID: x.EventID, UserID: x.UserID, Status: x.Status.Name(), Currency: x.Currency,
		Subtotal: x.Subtotal, Discount: x.Discount, Total: x.Total, PromoCode: x.PromoCode, BuyerName: x.BuyerName,
		BuyerEmail: x.BuyerEmail, BuyerPhone: x.BuyerPhone, SessionID: x.SessionID, ExpiresAt: x.ExpiresAt, PaymentMethod: string(x.PaymentMethod),
		PaidAt: x.PaidAt, RefundAmount: x.RefundAmount, StatusNote: x.StatusNote, CreatedAt: x.CreatedAt,
		Items: make([]orderItemResponse, 0, len(x.Items))}
	for _, it := range x.Items {
		out.Items = append(out.Items, orderItemResponse{TicketTypeID: it.TicketTypeID, Name: it.Name, Quantity: it.Quantity,
			UnitPrice: it.UnitPrice, SeatIDs: it.SeatIDs})
	}
	for _, t := range x.Tickets {
		out.Tickets = append(out.Tickets, toTicketResponse(t))
	}
	return out
}

type ticketResponse struct {
	ID            int64      `json:"id"`
	OrderID       int64      `json:"order_id"`
	EventID       int64      `json:"event_id"`
	EventTitle    string     `json:"event_title,omitempty"`
	EventStartsAt *time.Time `json:"event_starts_at,omitempty"` // when the ticket's showtime starts
	EventEndsAt   *time.Time `json:"event_ends_at,omitempty"`
	SessionID     int64      `json:"session_id,omitempty"`
	SessionLabel  string     `json:"session_label,omitempty"`
	Venue         string     `json:"venue,omitempty"`
	TicketTypeID  int64      `json:"ticket_type_id"`
	TypeName      string     `json:"type_name"`
	SeatID        int64      `json:"seat_id,omitempty"`
	Seat          string     `json:"seat,omitempty"`
	Code          string     `json:"code,omitempty"` // what the QR code encodes
	Status        string     `json:"status"`
	HolderID      int64      `json:"holder_id,omitempty"`
	HolderName    string     `json:"holder_name,omitempty"`
	HolderEmail   string     `json:"holder_email,omitempty"`
	TransferCount int        `json:"transfer_count"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
}

func toTicketResponse(t domain.Ticket) ticketResponse {
	r := ticketResponse{ID: t.ID, OrderID: t.OrderID, EventID: t.EventID, EventTitle: t.EventTitle, Venue: t.Venue,
		TicketTypeID: t.TicketTypeID, TypeName: t.TypeName, SeatID: t.SeatID, Seat: t.SeatLabel, Code: t.Code,
		Status: t.Status.Name(), HolderID: t.HolderID, HolderName: t.HolderName, HolderEmail: t.HolderEmail,
		TransferCount: t.TransferCount, CheckedInAt: t.CheckedInAt}
	if !t.EventStartsAt.IsZero() {
		r.EventStartsAt = &t.EventStartsAt
	}
	if !t.EventEndsAt.IsZero() {
		r.EventEndsAt = &t.EventEndsAt
	}
	r.SessionID, r.SessionLabel = t.SessionID, t.SessionLabel
	return r
}

type ticketEventResponse struct {
	ID        int64     `json:"id"`
	Event     string    `json:"event"`
	From      string    `json:"from_status,omitempty"`
	To        string    `json:"to_status,omitempty"`
	Actor     int64     `json:"actor,omitempty"` // 0 / absent: the system
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func toTicketEventResponse(e domain.TicketEvent) ticketEventResponse {
	r := ticketEventResponse{ID: e.ID, Event: e.Event, Actor: e.Actor, Note: e.Note, CreatedAt: e.CreatedAt}
	if e.FromStatus != 0 {
		r.From = e.FromStatus.Name()
	}
	if e.ToStatus != 0 {
		r.To = e.ToStatus.Name()
	}
	return r
}

type holderRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type transferRequest struct {
	ToUserID int64  `json:"to_user_id"`
	ToEmail  string `json:"to_email"`
	Message  string `json:"message"`
}

type transferResponse struct {
	ID         int64           `json:"id"`
	TicketID   int64           `json:"ticket_id"`
	FromUser   int64           `json:"from_user"`
	ToUser     int64           `json:"to_user,omitempty"`
	ToEmail    string          `json:"to_email,omitempty"`
	Status     string          `json:"status"`
	Message    string          `json:"message,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  time.Time       `json:"expires_at"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
	Ticket     *ticketResponse `json:"ticket,omitempty"`
}

func toTransferResponse(t domain.Transfer) transferResponse {
	r := transferResponse{ID: t.ID, TicketID: t.TicketID, FromUser: t.FromUser, ToUser: t.ToUser, ToEmail: t.ToEmail,
		Status: t.Status.Name(), Message: t.Message, CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt, ResolvedAt: t.ResolvedAt}
	if t.Ticket != nil {
		tk := toTicketResponse(*t.Ticket)
		tk.Code, tk.HolderEmail = "", ""
		r.Ticket = &tk
	}
	return r
}

type staffRequest struct {
	UserID int64 `json:"user_id"`
}

type staffResponse struct {
	UserID    int64     `json:"user_id"`
	AddedBy   int64     `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
}

type waitlistResponse struct {
	TicketTypeID int64     `json:"ticket_type_id"`
	EventID      int64     `json:"event_id"`
	EventTitle   string    `json:"event_title"`
	TypeName     string    `json:"type_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type scanRequest struct {
	Scans []struct {
		Code      string     `json:"code"`
		ScannedAt *time.Time `json:"scanned_at"`
	} `json:"scans"`
}

type scanResponse struct {
	Code   string          `json:"code"`
	Result string          `json:"result"`
	Ticket *ticketResponse `json:"ticket,omitempty"`
}

type inviteRequest struct {
	TicketTypeID int64   `json:"ticket_type_id"`
	Quantity     int     `json:"quantity"`
	SeatIDs      []int64 `json:"seat_ids"`
	UserID       int64   `json:"user_id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Note         string  `json:"note"`
}

type attendeeResponse struct {
	ticketResponse
	BuyerName  string `json:"buyer_name"`
	BuyerEmail string `json:"buyer_email"`
}

// ---- reports, notifications

type reportResponse struct {
	EventID     int64                `json:"event_id"`
	Orders      int                  `json:"orders"`
	TicketsSold int                  `json:"tickets_sold"`
	Revenue     int64                `json:"revenue"`
	Refunded    int64                `json:"refunded"`
	CheckedIn   int                  `json:"checked_in"`
	Invited     int                  `json:"invited"`
	ByType      []typeSalesResponse  `json:"by_type"`
	Daily       []dailySalesResponse `json:"daily"`
}

type dailySalesResponse struct {
	Day     string `json:"day"`
	Orders  int    `json:"orders"`
	Tickets int    `json:"tickets"`
	Revenue int64  `json:"revenue"`
}

type typeSalesResponse struct {
	TicketTypeID int64  `json:"ticket_type_id"`
	Name         string `json:"name"`
	Total        int    `json:"total"`
	Sold         int    `json:"sold"`
	Held         int    `json:"held"`
	Revenue      int64  `json:"revenue"`
}

func toReportResponse(r domain.EventReport) reportResponse {
	out := reportResponse{EventID: r.EventID, Orders: r.Orders, TicketsSold: r.TicketsSold, Revenue: r.Revenue,
		Refunded: r.Refunded, CheckedIn: r.CheckedIn, Invited: r.Invited, ByType: make([]typeSalesResponse, 0, len(r.ByType)),
		Daily: make([]dailySalesResponse, 0, len(r.Daily))}
	for _, x := range r.Daily {
		out.Daily = append(out.Daily, dailySalesResponse{Day: x.Day.Format("2006-01-02"), Orders: x.Orders, Tickets: x.Tickets, Revenue: x.Revenue})
	}
	for _, s := range r.ByType {
		out.ByType = append(out.ByType, typeSalesResponse{TicketTypeID: s.TicketTypeID, Name: s.Name, Total: s.Total, Sold: s.Sold, Held: s.Held, Revenue: s.Revenue})
	}
	return out
}

type notificationResponse struct {
	ID        int64      `json:"id"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	OrderID   int64      `json:"order_id,omitempty"`
	EventID   int64      `json:"event_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// ---- lists

type pageResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func toPage[S, T any](p domain.Page[S], conv func(S) T) pageResponse[T] {
	out := pageResponse[T]{Items: make([]T, 0, len(p.Items)), NextCursor: encodeIDCursor(p.NextID)}
	for _, x := range p.Items {
		out.Items = append(out.Items, conv(x))
	}
	return out
}

type seatMapRequest struct {
	domain.SeatMapLayout
	Seats []struct {
		Row    string  `json:"row"`
		Number int     `json:"number"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
	} `json:"seats"`
}

type duplicateRequest struct {
	Title          string    `json:"title"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	CopyPromotions bool      `json:"copy_promotions"`
}

type resaleRequest struct {
	Price int64 `json:"price"`
}

type resaleResponse struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	EventID   int64     `json:"event_id"`
	TypeName  string    `json:"type_name"`
	Seat      string    `json:"seat,omitempty"`
	Price     int64     `json:"price"`
	FacePrice int64     `json:"face_price"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// the seller sees these; the public does not
	SellerID int64 `json:"seller_id,omitempty"`
	Fee      int64 `json:"fee,omitempty"`
}

func toResaleResponse(l domain.ResaleListing, seller bool) resaleResponse {
	r := resaleResponse{ID: l.ID, TicketID: l.TicketID, EventID: l.EventID, TypeName: l.TypeName, Seat: l.SeatLabel, Price: l.Price,
		FacePrice: l.FacePrice, Currency: l.Currency, Status: l.Status.Name(), CreatedAt: l.CreatedAt}
	if seller {
		r.SellerID, r.Fee = l.SellerID, l.Fee
	}
	return r
}

// ---- venues

type venueRequest struct {
	Name        string           `json:"name"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
	Description string           `json:"description"`
	Template    string           `json:"template"`
	Params      map[string]int   `json:"params"`
	Map         *domain.VenueMap `json:"map"`
}

func (r venueRequest) input() inbound.VenueInput {
	return inbound.VenueInput{Name: r.Name, City: r.City, Address: r.Address, Description: r.Description, Template: r.Template, Params: r.Params, Map: r.Map}
}

type venueResponse struct {
	ID          int64            `json:"id"`
	OwnerID     int64            `json:"owner_id"`
	Name        string           `json:"name"`
	City        string           `json:"city,omitempty"`
	Address     string           `json:"address,omitempty"`
	Description string           `json:"description,omitempty"`
	Shared      bool             `json:"shared"`
	Seats       int              `json:"seats"`
	Sections    int              `json:"sections"`
	CreatedAt   time.Time        `json:"created_at"`
	Map         *domain.VenueMap `json:"map,omitempty"`
}

func toVenueResponse(v domain.Venue, withMap bool) venueResponse {
	r := venueResponse{ID: v.ID, OwnerID: v.OwnerID, Name: v.Name, City: v.City, Address: v.Address, Description: v.Description,
		Shared: v.Shared, Seats: v.Seats, Sections: v.Sections, CreatedAt: v.CreatedAt}
	if withMap {
		r.Map = &v.Map
	}
	return r
}

type sectionCountResponse struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Seats      int    `json:"seats"`
	Accessible int    `json:"accessible"`
}

func toSectionCounts(in []inbound.SectionCount) []sectionCountResponse {
	out := make([]sectionCountResponse, len(in))
	for i, c := range in {
		out[i] = sectionCountResponse{Key: c.Key, Name: c.Name, Seats: c.Seats, Accessible: c.Accessible}
	}
	return out
}

type generatedSeatResponse struct {
	Section    string  `json:"section"`
	Row        string  `json:"row"`
	Number     int     `json:"number"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Accessible bool    `json:"accessible,omitempty"`
}

func toPreviewResponse(p inbound.VenuePreview, withSeats bool) map[string]any {
	out := map[string]any{"total": p.Total, "sections": toSectionCounts(p.Sections), "layout": p.View}
	if withSeats {
		seats := make([]generatedSeatResponse, len(p.Seats))
		for i, s := range p.Seats {
			seats[i] = generatedSeatResponse{Section: s.Section, Row: s.Row, Number: s.Number, X: s.X, Y: s.Y, Accessible: s.Accessible}
		}
		out["seats"] = seats
	}
	return out
}

type templateResponse struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      []templateParam `json:"params"`
}

type templateParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     int    `json:"default"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
}

type applyVenueRequest struct {
	VenueID     int64 `json:"venue_id"`
	Assignments []struct {
		Section      string `json:"section"`
		TicketTypeID int64  `json:"ticket_type_id"`
	} `json:"assignments"`
	Replace bool `json:"replace"`
}

func timeNow() time.Time { return time.Now() }

// ---- sessions

type sessionRequest struct {
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
	Label    string    `json:"label"`
	CopyFrom int64     `json:"copy_from"`
}

type sessionResponse struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Label        string    `json:"label,omitempty"`
	Status       string    `json:"status"`
	CancelReason string    `json:"cancel_reason,omitempty"`
}

func toSessionResponse(s domain.Session) sessionResponse {
	return sessionResponse{ID: s.ID, EventID: s.EventID, StartsAt: s.StartsAt, EndsAt: s.EndsAt, Label: s.Label, Status: s.Status.Name(), CancelReason: s.CancelReason}
}
