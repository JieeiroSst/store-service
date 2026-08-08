package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/bot-service/config"
	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FacebookHandler struct {
	usecase port.BotUsecase
	cfg     config.FacebookConfig
}

func NewFacebookHandler(usecase port.BotUsecase, cfg *config.Config) *FacebookHandler {
	return &FacebookHandler{usecase: usecase, cfg: cfg.Facebook}
}

func (h *FacebookHandler) VerifyWebhook(c *gin.Context) {
	if c.Query("hub.verify_token") != h.cfg.VerifyToken {
		c.Status(http.StatusForbidden)
		return
	}
	c.String(http.StatusOK, c.Query("hub.challenge"))
}

type facebookWebhookPayload struct {
	Entry []struct {
		Messaging []struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Message struct {
				Text string `json:"text"`
				MID  string `json:"mid"`
			} `json:"message"`
		} `json:"messaging"`
	} `json:"entry"`
}

func (h *FacebookHandler) ReceiveWebhook(c *gin.Context) {
	lg := logger.WithContext(c.Request.Context())

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if h.cfg.AppSecret != "" && !validSignature(h.cfg.AppSecret, c.GetHeader("X-Hub-Signature-256"), body) {
		c.Status(http.StatusUnauthorized)
		return
	}

	var payload facebookWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, m := range entry.Messaging {
			if m.Message.Text == "" {
				continue
			}
			msg := model.IncomingMessage{
				Channel:   model.ChannelFacebook,
				ChatID:    m.Sender.ID,
				MessageID: m.Message.MID,
				Text:      m.Message.Text,
			}
			if err := h.usecase.HandleMessage(c.Request.Context(), msg); err != nil {
				lg.Error("facebook webhook: handle message", zap.Error(err))
			}
		}
	}
	c.Status(http.StatusOK)
}

func validSignature(appSecret, header string, body []byte) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimPrefix(header, prefix)))
}
