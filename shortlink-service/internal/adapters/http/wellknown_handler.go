package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// WellKnownHandler mirrors src/routes/well-known.ts.
type WellKnownHandler struct {
	iosTeamID                 string
	iosBundleID               string
	androidPackageName        string
	androidSHA256Fingerprints string
}

func NewWellKnownHandler(iosTeamID, iosBundleID, androidPackageName, androidSHA256Fingerprints string) *WellKnownHandler {
	return &WellKnownHandler{iosTeamID, iosBundleID, androidPackageName, androidSHA256Fingerprints}
}

// AppleAppSiteAssociation mirrors GET /.well-known/apple-app-site-association.
func (h *WellKnownHandler) AppleAppSiteAssociation(c *gin.Context) {
	if h.iosTeamID == "" || h.iosBundleID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration missing",
			"message": "iOS app configuration not found. Please set IOS_TEAM_ID and IOS_BUNDLE_ID environment variables.",
			"docs":    "https://docs.linkforty.com/guides/sdk-integration#ios-universal-links",
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"applinks": gin.H{
			"apps": []string{},
			"details": []gin.H{
				{
					"appID": h.iosTeamID + "." + h.iosBundleID,
					"paths": []string{"*"},
				},
			},
		},
	})
}

// AssetLinks mirrors GET /.well-known/assetlinks.json.
func (h *WellKnownHandler) AssetLinks(c *gin.Context) {
	if h.androidPackageName == "" || h.androidSHA256Fingerprints == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration missing",
			"message": "Android app configuration not found. Please set ANDROID_PACKAGE_NAME and ANDROID_SHA256_FINGERPRINTS environment variables.",
			"docs":    "https://docs.linkforty.com/guides/sdk-integration#android-app-links",
		})
		return
	}

	fingerprints := make([]string, 0)
	for _, fp := range strings.Split(h.androidSHA256Fingerprints, ",") {
		fp = strings.TrimSpace(fp)
		if fp != "" {
			fingerprints = append(fingerprints, fp)
		}
	}

	if len(fingerprints) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid configuration",
			"message": "ANDROID_SHA256_FINGERPRINTS is empty or invalid. Must be comma-separated list of SHA-256 fingerprints.",
			"example": "ANDROID_SHA256_FINGERPRINTS=AA:BB:CC:DD:EE:FF:...,11:22:33:44:55:66:...",
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, []gin.H{
		{
			"relation": []string{"delegate_permission/common.handle_all_urls"},
			"target": gin.H{
				"namespace":                "android_app",
				"package_name":             h.androidPackageName,
				"sha256_cert_fingerprints": fingerprints,
			},
		},
	})
}
