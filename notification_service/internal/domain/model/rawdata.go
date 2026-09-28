package model

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
)

const (
	MaxRawDataBytes   = 64 << 10
	maxSlackTextChars = 35000
)

func RawDataSize(data map[string]any) int {
	if len(data) == 0 {
		return 0
	}
	b, _ := json.Marshal(data)
	return len(b)
}

func sortedKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func rawValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		return string(b)
	}
}

func RawDataHTML(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<table style="border-collapse:collapse;margin-top:16px;font-family:Arial,sans-serif;font-size:14px">`)
	for _, k := range sortedKeys(data) {
		fmt.Fprintf(&b, `<tr><th style="text-align:left;padding:6px 12px;border:1px solid #ddd;background:#f6f6f6">%s</th><td style="padding:6px 12px;border:1px solid #ddd;white-space:pre-wrap">%s</td></tr>`,
			html.EscapeString(k), html.EscapeString(rawValue(data[k])))
	}
	b.WriteString(`</table>`)
	return b.String()
}

func RawDataSlack(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ""
	}
	text := strings.ReplaceAll(string(pretty), "```", "'''")
	if len(text) > maxSlackTextChars {
		text = text[:maxSlackTextChars] + "\n…(truncated)"
	}
	return "```\n" + text + "\n```"
}

func TextToHTML(text string) string {
	return `<pre style="font-family:Arial,sans-serif;white-space:pre-wrap">` + html.EscapeString(text) + `</pre>`
}
