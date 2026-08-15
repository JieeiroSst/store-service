package voucher

// Status models the voucher lifecycle:
//
//	CREATED -> ISSUED -> ACTIVE -> REDEEMED
//	                            -> EXPIRED
//	                            -> REVOKED
//	CREATED -> REVOKED
//	ISSUED  -> REVOKED
type Status string

const (
	StatusCreated  Status = "created"
	StatusIssued   Status = "issued"
	StatusActive   Status = "active"
	StatusRedeemed Status = "redeemed"
	StatusExpired  Status = "expired"
	StatusRevoked  Status = "revoked"
)

func (s Status) Valid() bool {
	switch s {
	case StatusCreated, StatusIssued, StatusActive, StatusRedeemed, StatusExpired, StatusRevoked:
		return true
	default:
		return false
	}
}

var transitions = map[Status][]Status{
	StatusCreated:  {StatusIssued, StatusRevoked},
	StatusIssued:   {StatusActive, StatusExpired, StatusRevoked},
	StatusActive:   {StatusRedeemed, StatusExpired, StatusRevoked},
	StatusRedeemed: {},
	StatusExpired:  {},
	StatusRevoked:  {},
}

func (s Status) canTransitionTo(target Status) bool {
	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}
