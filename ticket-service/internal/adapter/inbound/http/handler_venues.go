package http

import (
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func (h *Handler) venueTemplates(w http.ResponseWriter, _ *http.Request) {
	ts := h.venues.Templates()
	out := make([]templateResponse, len(ts))
	for i, t := range ts {
		out[i] = templateResponse{Name: t.Name, Description: t.Description, Params: make([]templateParam, len(t.Params))}
		for j, p := range t.Params {
			out[i].Params[j] = templateParam{Name: p.Name, Description: p.Description, Default: p.Default, Min: p.Min, Max: p.Max}
		}
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) previewVenue(w http.ResponseWriter, r *http.Request) {
	var req venueRequest
	if !h.decode(w, r, &req) {
		return
	}
	p, err := h.venues.Preview(r.Context(), req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	out := toPreviewResponse(p, r.URL.Query().Get("seats") == "true")
	if req.Template != "" && req.Map == nil {
		out["map"] = p.Map // the blueprint a template stands for, so an editor can start from it
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) createVenue(w http.ResponseWriter, r *http.Request) {
	var req venueRequest
	if !h.decode(w, r, &req) {
		return
	}
	v, err := h.venues.Create(r.Context(), mustPrincipal(r), req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toVenueResponse(v, true))
}

func (h *Handler) getVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	v, err := h.venues.Get(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := map[string]any{"venue": toVenueResponse(v, true)}
	if r.URL.Query().Get("seats") == "true" {
		p, err := h.venues.Preview(r.Context(), inbound.VenueInput{Map: &v.Map})
		if err != nil {
			h.fail(w, err)
			return
		}
		out["preview"] = toPreviewResponse(p, true)
	}
	h.write(w, http.StatusOK, out)
}

func (h *Handler) updateVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req venueRequest
	if !h.decode(w, r, &req) {
		return
	}
	v, err := h.venues.Update(r.Context(), mustPrincipal(r), id, req.input())
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toVenueResponse(v, true))
}

func (h *Handler) deleteVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.venues.Delete(r.Context(), mustPrincipal(r), id); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listVenues(w http.ResponseWriter, r *http.Request) {
	after, ok := h.cursor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	if s := q.Get("scope"); s != "" && s != "mine" && s != "shared" {
		h.fail(w, fmt.Errorf("%w: scope must be mine or shared", domain.ErrInvalid))
		return
	}
	p, err := h.venues.List(r.Context(), mustPrincipal(r), q.Get("scope") == "shared", after, atoi(q.Get("limit")))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toPage(p, func(v domain.Venue) venueResponse { return toVenueResponse(v, false) }))
}

func (h *Handler) copyVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if r.ContentLength != 0 && !h.decode(w, r, &req) {
		return
	}
	v, err := h.venues.Copy(r.Context(), mustPrincipal(r), id, req.Name)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusCreated, toVenueResponse(v, true))
}

func (h *Handler) shareVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req struct {
		Shared bool `json:"shared"`
	}
	if !h.decode(w, r, &req) {
		return
	}
	v, err := h.venues.SetShared(r.Context(), mustPrincipal(r), id, req.Shared)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, toVenueResponse(v, false))
}

func (h *Handler) applyVenue(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	var req applyVenueRequest
	if !h.decode(w, r, &req) {
		return
	}
	in := inbound.ApplyVenueInput{VenueID: req.VenueID, Replace: req.Replace}
	for _, a := range req.Assignments {
		in.Assignments = append(in.Assignments, inbound.SectionAssignment{Section: a.Section, TicketTypeID: a.TicketTypeID})
	}
	res, err := h.events.ApplyVenueMap(r.Context(), mustPrincipal(r), id, in)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.write(w, http.StatusOK, map[string]any{"seats": res.Seats, "sections": toSectionCounts(res.Sections)})
}

func (h *Handler) eventSeatMap(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	m, err := h.catalog.EventSeatMap(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	seats := make([]seatResponse, len(m.Seats))
	for i, s := range m.Seats {
		seats[i] = toSeatResponse(s)
	}
	types := make([]ticketTypeResponse, len(m.Types))
	now := timeNow()
	for i, t := range m.Types {
		types[i] = toTypeResponse(t, now)
	}
	h.write(w, http.StatusOK, map[string]any{"layout": m.Layout, "ticket_types": types, "seats": seats})
}
