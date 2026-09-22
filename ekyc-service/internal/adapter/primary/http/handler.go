package http

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type Handler struct {
	usecase port.EkycUsecase
}

func NewHandler(usecase port.EkycUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) SubmitCitizenCard(c *gin.Context) {
	userID := c.Param("user_id")

	front, err := readFormFile(c, "front")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or unreadable 'front' image: " + err.Error()})
		return
	}
	back, err := readFormFile(c, "back")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or unreadable 'back' image: " + err.Error()})
		return
	}

	identity, err := h.usecase.SubmitCitizenCard(c.Request.Context(), userID, front, back)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, identity)
}

func (h *Handler) SubmitFaceScan(c *gin.Context) {
	userID := c.Param("user_id")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected multipart form with 'frames' file(s): " + err.Error()})
		return
	}
	fileHeaders := form.File["frames"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one 'frames' image is required"})
		return
	}

	frames := make([][]byte, 0, len(fileHeaders))
	for _, fh := range fileHeaders {
		data, err := readFileHeader(fh)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable frame: " + err.Error()})
			return
		}
		frames = append(frames, data)
	}

	face, err := h.usecase.SubmitFaceScan(c.Request.Context(), userID, frames)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, face)
}

func (h *Handler) SubmitNFCChip(c *gin.Context) {
	userID := c.Param("user_id")

	dump := port.ChipDump{}
	var err error

	if dump.EFCOM, err = readOptionalFormFile(c, "ef_com"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable 'ef_com': " + err.Error()})
		return
	}
	if dump.DG1, err = readOptionalFormFile(c, "dg1"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable 'dg1': " + err.Error()})
		return
	}
	if dump.DG2, err = readOptionalFormFile(c, "dg2"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable 'dg2': " + err.Error()})
		return
	}
	if dump.EFSOD, err = readOptionalFormFile(c, "ef_sod"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable 'ef_sod': " + err.Error()})
		return
	}
	if len(dump.EFCOM) == 0 && len(dump.DG1) == 0 && len(dump.DG2) == 0 && len(dump.EFSOD) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one of ef_com, dg1, dg2, ef_sod is required"})
		return
	}

	identity, err := h.usecase.SubmitNFCChip(c.Request.Context(), userID, dump)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, identity)
}

func (h *Handler) Verify(c *gin.Context) {
	userID := c.Param("user_id")
	v, err := h.usecase.Verify(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) GetStatus(c *gin.Context) {
	userID := c.Param("user_id")
	status, err := h.usecase.GetStatus(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func readFormFile(c *gin.Context, field string) ([]byte, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return nil, err
	}
	return readFileHeader(fh)
}

func readOptionalFormFile(c *gin.Context, field string) ([]byte, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return nil, nil
	}
	return readFileHeader(fh)
}

func readFileHeader(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrUserNotFound),
		errors.Is(err, port.ErrIdentityNotFound),
		errors.Is(err, port.ErrFaceNotFound),
		errors.Is(err, port.ErrVerificationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrNoFaceDetected), errors.Is(err, port.ErrMRZNotReadable):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
