package http

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/polymarket-service/config"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Handler struct {
	catalog    port.CatalogUsecase
	resolution port.ResolutionUsecase
	exchange   port.ExchangeUsecase
	accounts   port.AccountUsecase
	social     port.SocialUsecase
	referral   port.ReferralUsecase
	rewards    port.RewardUsecase
	identity   port.IdentityUsecase
	stream     port.EventStream
	adminToken string
	tokenAuth  bool
}

type Params struct {
	fx.In

	Catalog    port.CatalogUsecase
	Resolution port.ResolutionUsecase
	Exchange   port.ExchangeUsecase
	Accounts   port.AccountUsecase
	Social     port.SocialUsecase
	Referral   port.ReferralUsecase
	Rewards    port.RewardUsecase
	Identity   port.IdentityUsecase
	Stream     port.EventStream
	Config     *config.Config
}

func NewHandler(p Params) *Handler {
	return &Handler{
		catalog: p.Catalog, resolution: p.Resolution, exchange: p.Exchange, accounts: p.Accounts,
		social: p.Social, referral: p.Referral, rewards: p.Rewards, identity: p.Identity, stream: p.Stream,
		adminToken: p.Config.Admin.Token, tokenAuth: p.Config.UserService.AuthMode == "token",
	}
}

func (h *Handler) requireAdmin(c *gin.Context) {
	got := c.GetHeader("X-Admin-Token")
	if h.adminToken == "" || subtle.ConstantTimeCompare([]byte(got), []byte(h.adminToken)) != 1 {
		writeError(c, port.ErrUnauthorized)
		c.Abort()
	}
}

const uidKey = "uid"

// authenticate runs on every API route. In token mode it turns a bearer token
// into a user id through user-service (a token that is present but not valid
// is refused outright). In gateway mode it does nothing: the gateway has
// already authenticated the caller and sets X-User-Id.
func (h *Handler) authenticate(c *gin.Context) {
	if !h.tokenAuth {
		return
	}
	header := c.GetHeader("Authorization")
	if header == "" {
		return
	}
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		writeError(c, port.ErrUnauthenticated)
		c.Abort()
		return
	}
	id, err := h.identity.Authenticate(c.Request.Context(), strings.TrimSpace(token))
	if err != nil {
		writeError(c, err)
		c.Abort()
		return
	}
	c.Set(uidKey, id)
}

// actor is the authenticated user, if any. In token mode X-User-Id is ignored:
// only user-service can vouch for a user.
func (h *Handler) actor(c *gin.Context) string {
	if h.tokenAuth {
		return c.GetString(uidKey)
	}
	return c.GetHeader("X-User-Id")
}

// requireOwner guards /users/:userId routes: the authenticated user must be the
// one in the path. In gateway mode an absent identity is trusted (internal
// calls); in token mode it is refused.
func (h *Handler) requireOwner(c *gin.Context) {
	id := h.actor(c)
	switch {
	case id == "" && h.tokenAuth:
		writeError(c, port.ErrUnauthenticated)
		c.Abort()
	case id != "" && id != c.Param("userId"):
		writeError(c, port.ErrForbidden)
		c.Abort()
	}
}

// caller resolves the acting user for identities carried in a body or query:
// the authenticated identity wins, and a body naming someone else is refused.
func (h *Handler) caller(c *gin.Context, bodyUserID string) (string, bool) {
	id := h.actor(c)
	switch {
	case id == "" && h.tokenAuth:
		writeError(c, port.ErrUnauthenticated)
		return "", false
	case id != "" && bodyUserID != "" && id != bodyUserID:
		writeError(c, port.ErrForbidden)
		return "", false
	case id != "":
		return id, true
	}
	return bodyUserID, true
}

func bind(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func pathInt(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, name string) int {
	n, _ := strconv.Atoi(c.Query(name))
	return n
}

func outcomeParam(c *gin.Context, def model.Outcome) model.Outcome {
	if v := c.Query("outcome"); v != "" {
		return model.Outcome(v)
	}
	return def
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrUnauthorized), errors.Is(err, port.ErrUnauthenticated):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrNotFound), errors.Is(err, port.ErrWalletNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrAlreadyExists), errors.Is(err, port.ErrReferralNotAllowed),
		errors.Is(err, port.ErrNegRisk):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInvalidInput),
		errors.Is(err, port.ErrInvalidUser),
		errors.Is(err, port.ErrInvalidOutcome),
		errors.Is(err, port.ErrInvalidPrice),
		errors.Is(err, port.ErrInvalidSize),
		errors.Is(err, port.ErrNotNegRisk),
		errors.Is(err, port.ErrWalletRejected):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrMarketNotTradable),
		errors.Is(err, port.ErrOrderNotOpen),
		errors.Is(err, port.ErrNotFilled),
		errors.Is(err, port.ErrInvalidTransition),
		errors.Is(err, port.ErrDisputeWindowOpen),
		errors.Is(err, port.ErrDisputeWindowShut):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInsufficientBalance),
		errors.Is(err, port.ErrInsufficientShares),
		errors.Is(err, port.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrWalletUnavailable), errors.Is(err, port.ErrUpstream):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
