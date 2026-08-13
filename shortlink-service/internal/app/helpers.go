package app

import (
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"gorm.io/datatypes"
)

func parseISOTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func mustMapJSON(m map[string]interface{}) datatypes.JSON {
	if m == nil {
		m = map[string]interface{}{}
	}
	b, _ := json.Marshal(m)
	return datatypes.JSON(b)
}

func mustUTMJSON(u domain.UTMParameters) datatypes.JSON {
	b, _ := json.Marshal(u)
	return datatypes.JSON(b)
}

func mustTargetingJSON(t domain.TargetingRules) datatypes.JSON {
	b, _ := json.Marshal(t)
	return datatypes.JSON(b)
}

func linkTemplateSettingsToJSONForPatch(s domain.LinkTemplateSettings) datatypes.JSON {
	b, _ := json.Marshal(s)
	return datatypes.JSON(b)
}
