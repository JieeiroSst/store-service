package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/JieeiroSst/banking-service/pkg/authn"
)

func ownerParam(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !principal(c).Owns(c.Param(param)) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func ownerQuery(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !principal(c).Owns(c.Query(field)) {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		c.Next()
	}
}

func forceOwnerBody(field string, numeric bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := principal(c)
		if p.IsAdmin() {
			c.Next()
			return
		}
		if !numeric {
			setBodyField(c, field, p.UserID)
			c.Next()
			return
		}
		id, err := strconv.ParseInt(p.UserID, 10, 64)
		if err != nil {
			abortAuth(c, authn.ErrForbidden)
			return
		}
		setBodyField(c, field, id)
		c.Next()
	}
}
