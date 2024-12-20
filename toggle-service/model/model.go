package model

type Countries struct {
	CountryId       int    `json:"country_id,omitempty" gorm:"primaryKey"`
	CountryName     string `json:"country_name,omitempty"`
	DefaultLanguage string `json:"default_language,omitempty"`
	Currency        string `json:"currency,omitempty"`
	Timezone        string `json:"timezone,omitempty"`
}

type Users struct {
	UserId    int       `json:"user_id,omitempty" gorm:"primaryKey"`
	UserName  string    `json:"user_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CountryId int       `json:"country_id,omitempty"`
	Countries Countries `json:"countries,omitempty" gorm:"foreignKey:CountryId;references:CountryId"`
}

type Features struct {
	FeatureId   int    `json:"feature_id,omitempty" gorm:"primaryKey"`
	FeatureName string `json:"feature_name,omitempty"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
}

type FeatureCountries struct {
	FeatureId int         `json:"feature_id,omitempty" gorm:"primaryKey"`
	CountryId int         `json:"country_id,omitempty" gorm:"primaryKey"`
	Features  []Features  `json:"features,omitempty" gorm:"foreignkey:FeatureId"`
	Countries []Countries `json:"countries,omitempty" gorm:"foreignkey:CountryId"`
}
