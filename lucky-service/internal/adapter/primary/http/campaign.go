package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

type publicPrize struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	Type     model.PrizeType `json:"type"`
	Points   int64           `json:"points"`
	Position int             `json:"position"`
	ImageURL string          `json:"image_url"`
}

type publicCampaign struct {
	ID             int64           `json:"id"`
	GameType       model.GameType  `json:"game_type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	StartAt        time.Time       `json:"start_at"`
	EndAt          time.Time       `json:"end_at"`
	Timezone       string          `json:"timezone"`
	DailySpins     int             `json:"daily_spins"`
	Events         []string        `json:"extra_play_events"`
	CheckinRewards []int64         `json:"checkin_rewards"`
	Settings       json.RawMessage `json:"settings"`
	Prizes         []publicPrize   `json:"prizes"`
}

func public(c *model.Campaign) publicCampaign {
	out := publicCampaign{
		ID:             c.ID,
		GameType:       c.GameType,
		Name:           c.Name,
		Description:    c.Description,
		StartAt:        c.StartAt,
		EndAt:          c.EndAt,
		Timezone:       c.Timezone,
		DailySpins:     c.DailySpins,
		Events:         make([]string, 0, len(c.SpinRules)),
		CheckinRewards: c.CheckinRewards,
		Settings:       c.Settings,
		Prizes:         make([]publicPrize, 0, len(c.Prizes)),
	}
	for _, r := range c.SpinRules {
		out.Events = append(out.Events, r.Event)
	}
	for _, p := range c.Prizes {
		out.Prizes = append(out.Prizes, publicPrize{ID: p.ID, Name: p.Name, Type: p.Type, Points: p.Points, Position: p.Position, ImageURL: p.ImageURL})
	}
	return out
}

func privileged(r *http.Request) bool {
	a, ok := actorFrom(r)
	return ok && (a.Admin || a.Service)
}

func (h *Handler) Lobby(w http.ResponseWriter, r *http.Request) {
	list, err := h.campaigns.Lobby(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	if privileged(r) {
		writeData(w, http.StatusOK, list)
		return
	}
	out := make([]publicCampaign, 0, len(list))
	for i := range list {
		out = append(out, public(&list[i]))
	}
	writeData(w, http.StatusOK, out)
}

func (h *Handler) CurrentCampaign(w http.ResponseWriter, r *http.Request) {
	c, err := h.campaigns.Current(r.Context(), r.PathValue("game"))
	if err != nil {
		writeError(w, err)
		return
	}
	if privileged(r) {
		writeData(w, http.StatusOK, c)
		return
	}
	writeData(w, http.StatusOK, public(c))
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var in port.CampaignInput
	if !decode(w, r, &in) {
		return
	}
	c, err := h.campaigns.Create(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, c)
}

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	p, err := h.campaigns.List(r.Context(), actor, port.ListCampaignsInput{GameType: q.Get("game_type"), Status: q.Get("status"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	c, err := h.campaigns.Get(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, c)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in port.CampaignInput
	if !decode(w, r, &in) {
		return
	}
	c, err := h.campaigns.Update(r.Context(), actor, id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, c)
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.campaigns.Delete(r.Context(), actor, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ActivateCampaign(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.campaigns.Activate)
}

func (h *Handler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.campaigns.Pause)
}

func (h *Handler) transition(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error)) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	c, err := fn(r.Context(), actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, c)
}

func (h *Handler) CampaignPlays(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	p, err := h.games.CampaignPlays(r.Context(), actor, port.ListPlaysInput{CampaignID: id, UserID: r.URL.Query().Get("user_id"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}
