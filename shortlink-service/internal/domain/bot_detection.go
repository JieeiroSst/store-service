package domain

import "strings"

type BotReason string

const (
	BotReasonEdge   BotReason = "edge"
	BotReasonMethod BotReason = "method"
	BotReasonUA     BotReason = "ua"
)

type BotClassification struct {
	IsBot  bool
	Reason BotReason // empty when IsBot is false
}

var nonHumanMethods = map[string]struct{}{
	"HEAD":    {},
	"OPTIONS": {},
}

var botUAPatterns = []string{
	"bot", "crawl", "spider", "slurp", "search", "fetch", "monitor",
	"facebookexternalhit", "facebot", "twitterbot", "linkedinbot",
	"slackbot", "discordbot", "telegrambot", "whatsapp", "pinterestbot",
	"skypeuripreview", "googlebot", "bingbot", "ia_archiver", "yandex",
	"duckduckbot", "baiduspider", "applebot", "semrushbot", "ahrefsbot",
	"mj12bot", "dotbot", "curl", "wget", "python-requests", "headlesschrome",
	"phantomjs", "postmanruntime", "go-http-client", "okhttp", "axios",
}

func ClassifyBot(userAgent, method string, edgeIsBot *bool) BotClassification {
	if edgeIsBot != nil && *edgeIsBot {
		return BotClassification{IsBot: true, Reason: BotReasonEdge}
	}
	if method != "" {
		if _, ok := nonHumanMethods[strings.ToUpper(method)]; ok {
			return BotClassification{IsBot: true, Reason: BotReasonMethod}
		}
	}
	lower := strings.ToLower(userAgent)
	for _, p := range botUAPatterns {
		if strings.Contains(lower, p) {
			return BotClassification{IsBot: true, Reason: BotReasonUA}
		}
	}
	return BotClassification{}
}

func EdgeBotSignal(trustEdgeBotHeader bool, headerValue string) *bool {
	if !trustEdgeBotHeader || headerValue == "" {
		return nil
	}
	v := headerValue == "1" || strings.EqualFold(headerValue, "true")
	return &v
}

var isSocialScraperPatterns = []string{
	"facebookexternalhit", "facebot", "twitterbot", "linkedinbot",
	"slackbot", "discordbot", "telegrambot", "whatsapp", "pinterestbot",
	"skypeuripreview", "googlebot", "bingbot", "ia_archiver",
}

func IsSocialScraper(userAgent string) bool {
	lower := strings.ToLower(userAgent)
	for _, p := range isSocialScraperPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
