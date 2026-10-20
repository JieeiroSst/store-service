package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
)

type playRequest struct {
	Choice int `json:"choice"`
}

func (h *Handler) MyStatus(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	st, err := h.games.Status(r.Context(), actor, r.PathValue("game"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, st)
}

func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var body playRequest
	if !decode(w, r, &body) {
		return
	}
	res, err := h.games.Play(r.Context(), actor, port.PlayInput{
		Game:           r.PathValue("game"),
		Choice:         body.Choice,
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		Client:         clientInfo(r),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	code := http.StatusCreated
	if res.Replayed {
		code = http.StatusOK
		w.Header().Set("Idempotent-Replayed", "true")
	}
	writeData(w, code, res)
}

func (h *Handler) Winners(w http.ResponseWriter, r *http.Request) {
	limit, err := queryInt(r, "limit")
	if err != nil {
		writeError(w, err)
		return
	}
	list, err := h.games.Winners(r.Context(), r.PathValue("game"), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, list)
}

func (h *Handler) MyPlays(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	limit, offset, ok := paging(w, r)
	if !ok {
		return
	}
	in := port.ListPlaysInput{GameType: r.URL.Query().Get("game_type"), Limit: limit, Offset: offset}
	if v := r.URL.Query().Get("campaign_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, errorBody("invalid_argument", "campaign_id: must be a positive integer"))
			return
		}
		in.CampaignID = id
	}
	p, err := h.games.History(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, p)
}

func (h *Handler) GrantPlays(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	var in port.GrantInput
	if !decode(w, r, &in) {
		return
	}
	res, err := h.games.GrantPlays(r.Context(), actor, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, res)
}
