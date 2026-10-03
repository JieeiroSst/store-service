package application

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/JIeeiroSst/webrtc-service/config"
	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/JIeeiroSst/webrtc-service/internal/port"
)

type ICEService struct {
	cfg config.ICEConfig
	now func() time.Time
}

func NewICEService(cfg *config.Config) port.ICEService {
	return &ICEService{cfg: cfg.ICE, now: time.Now}
}

func (s *ICEService) Servers(userID string) []domain.ICEServer {
	servers := []domain.ICEServer{}
	if len(s.cfg.STUNURLs) > 0 {
		servers = append(servers, domain.ICEServer{URLs: s.cfg.STUNURLs})
	}
	if len(s.cfg.TURNURLs) > 0 && s.cfg.TURNSecret != "" {
		if domain.ValidateID(userID) != nil {
			userID = "anonymous"
		}
		username := strconv.FormatInt(s.now().Add(s.cfg.TURNTTL).Unix(), 10) + ":" + userID
		mac := hmac.New(sha1.New, []byte(s.cfg.TURNSecret))
		mac.Write([]byte(username))
		servers = append(servers, domain.ICEServer{
			URLs:       s.cfg.TURNURLs,
			Username:   username,
			Credential: base64.StdEncoding.EncodeToString(mac.Sum(nil)),
		})
	}
	return servers
}
