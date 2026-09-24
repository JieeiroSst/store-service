package httpapi

import (
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
)

// DocumentHandler serves the files attached to a patient. Like the identity
// routes it is REST-only: file uploads are multipart, which the protobuf
// contract cannot express.
type DocumentHandler struct{ uc port.DocumentUsecase }

func NewDocumentHandler(uc port.DocumentUsecase) *DocumentHandler { return &DocumentHandler{uc: uc} }

func (h *DocumentHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/patients/{id}/documents", h.upload)
	mux.HandleFunc("GET /v1/patients/{id}/documents", h.list)
	mux.HandleFunc("GET /v1/patients/{id}/documents/{doc}/content", h.download)
	mux.HandleFunc("DELETE /v1/patients/{id}/documents/{doc}", h.remove)
}

type documentJSON struct {
	ID          string    `json:"id"`
	PatientID   int32     `json:"patient_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

func toJSON(d model.Document) documentJSON {
	return documentJSON{ID: d.ID, PatientID: d.PatientID, FileName: d.FileName, ContentType: d.ContentType, Size: d.Size, CreatedAt: d.CreatedAt}
}

func (h *DocumentHandler) upload(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	form, ok := parseForm(w, r)
	if !ok {
		return
	}
	fhs := form.File["file"]
	if len(fhs) == 0 {
		writeError(w, model.Invalid("multipart form field 'file' is required"))
		return
	}
	f, err := fhs[0].Open()
	if err != nil {
		writeError(w, model.Invalid("unreadable upload"))
		return
	}
	defer f.Close()

	doc, err := h.uc.Upload(r.Context(), id, port.DocumentUpload{FileName: fhs[0].Filename, Body: f})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toJSON(*doc))
}

func (h *DocumentHandler) list(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	docs, total, err := h.uc.List(r.Context(), id, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]documentJSON, len(docs))
	for i, d := range docs {
		items[i] = toJSON(d)
	}
	writeJSON(w, http.StatusOK, map[string]any{"documents": items, "total_count": total})
}

// download streams the file as an attachment. The type is the one upload-service
// detected from the bytes, and nosniff stops a browser second-guessing it.
func (h *DocumentHandler) download(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	d, err := h.uc.Open(r.Context(), id, r.PathValue("doc"))
	if err != nil {
		writeError(w, err)
		return
	}
	defer d.Body.Close()

	hd := w.Header()
	hd.Set("Content-Type", d.Document.ContentType)
	hd.Set("Content-Length", strconv.FormatInt(d.Document.Size, 10))
	hd.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": d.Document.FileName}))
	hd.Set("X-Content-Type-Options", "nosniff")
	hd.Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, d.Body); err != nil {
		logrus.WithError(err).WithField("document_id", d.Document.ID).Warn("download interrupted")
	}
}

func (h *DocumentHandler) remove(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	if err := h.uc.Delete(r.Context(), id, r.PathValue("doc")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
