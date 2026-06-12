package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type response struct {
	Data  any        `json:"data,omitempty"`
	Meta  *pageMeta  `json:"meta,omitempty"`
	Error *apiError  `json:"error,omitempty"`
}

type pageMeta struct {
	Count      int    `json:"count"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	codeValidation      = "VALIDATION_ERROR"
	codeInvalidField    = "INVALID_FIELD"
	codeNotFound        = "NOT_FOUND"
	codeSelfReferral    = "SELF_REFERRAL"
	codeAlreadyReferred = "ALREADY_REFERRED"
	codeLinkNotActive   = "LINK_NOT_ACTIVE"
	codeRateLimit       = "RATE_LIMIT_EXCEEDED"
	codeInternal        = "INTERNAL_ERROR"
)

func respOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, response{Data: data})
}

func respCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, response{Data: data})
}

func respList(c *gin.Context, data any, count int, nextCursor string) {
	c.JSON(http.StatusOK, response{
		Data: data,
		Meta: &pageMeta{Count: count, NextCursor: nextCursor},
	})
}

func respBadRequest(c *gin.Context, code, message string) {
	c.JSON(http.StatusBadRequest, response{Error: &apiError{Code: code, Message: message}})
}

func respNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, response{Error: &apiError{Code: codeNotFound, Message: message}})
}

func respUnprocessable(c *gin.Context, code, message string) {
	c.JSON(http.StatusUnprocessableEntity, response{Error: &apiError{Code: code, Message: message}})
}

func respTooManyRequests(c *gin.Context, code, message string) {
	c.JSON(http.StatusTooManyRequests, response{Error: &apiError{Code: code, Message: message}})
}

func respInternal(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response{Error: &apiError{Code: codeInternal, Message: "internal server error"}})
}
