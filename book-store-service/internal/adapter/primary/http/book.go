package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBook(c *gin.Context) {
	var book model.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.book.CreateBook(c.Request.Context(), &book)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.book.GetBook(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListBooks(c *gin.Context) {
	result, err := h.book.ListBooks(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var book model.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	book.ID = id
	result, err := h.book.UpdateBook(c.Request.Context(), &book)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.book.DeleteBook(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetHighestPriceBook returns the book with the highest price in the store.
func (h *Handler) GetHighestPriceBook(c *gin.Context) {
	result, err := h.book.HighestPriceBook(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetMostPurchasedBook returns the best-selling book by purchase_count.
func (h *Handler) GetMostPurchasedBook(c *gin.Context) {
	result, err := h.book.MostPurchasedBook(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type purchaseRequest struct {
	Quantity int `json:"quantity"`
}

// PurchaseBook records a sale of the book, bumping purchase_count and
// decrementing quantity.
func (h *Handler) PurchaseBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req purchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	if err := h.book.PurchaseBook(c.Request.Context(), id, req.Quantity); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReadBook records a read of the book, bumping read_count.
func (h *Handler) ReadBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.book.ReadBook(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetCategoryPriceStats returns min/max/avg price and book count per category.
func (h *Handler) GetCategoryPriceStats(c *gin.Context) {
	result, err := h.book.CategoryPriceStats(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
