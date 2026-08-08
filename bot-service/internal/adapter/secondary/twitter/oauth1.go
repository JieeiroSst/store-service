package twitter

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// https://developer.x.com/en/docs/authentication/oauth-1-0a.
func signOAuth1(method, endpoint, consumerKey, consumerSecret, token, tokenSecret string) (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	params := map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            token,
		"oauth_version":          "1.0",
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	paramStr := strings.Join(pairs, "&")

	base := strings.ToUpper(method) + "&" + url.QueryEscape(endpoint) + "&" + url.QueryEscape(paramStr)
	signingKey := url.QueryEscape(consumerSecret) + "&" + url.QueryEscape(tokenSecret)

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(base))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	var b strings.Builder
	b.WriteString("OAuth ")
	b.WriteString(`oauth_consumer_key="` + url.QueryEscape(consumerKey) + `", `)
	b.WriteString(`oauth_nonce="` + url.QueryEscape(nonce) + `", `)
	b.WriteString(`oauth_signature="` + url.QueryEscape(signature) + `", `)
	b.WriteString(`oauth_signature_method="HMAC-SHA1", `)
	b.WriteString(`oauth_timestamp="` + timestamp + `", `)
	b.WriteString(`oauth_token="` + url.QueryEscape(token) + `", `)
	b.WriteString(`oauth_version="1.0"`)
	return b.String(), nil
}
