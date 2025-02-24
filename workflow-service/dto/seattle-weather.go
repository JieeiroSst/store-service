package dto

import (
	"github.com/JIeeiroSst/workflow-service/pkg/excel"
)

type SeattleWeatherRequestDTO struct {
	Date          string `json:"date" customFieldName:"date"`
	Precipitation string `json:"precipitation" customFieldName:"precipitation"`
	TempMax       string `json:"temp_max" customFieldName:"temp_max"`
	TempMin       string `json:"temp_min" customFieldName:"temp_min"`
	Wind          string `json:"wind" customFieldName:"wind"`
	Weather       string `json:"weather" customFieldName:"weather"`
	BatchID       string
}

func FormatSeattleWeather(generated []excel.AutoGenerated) (seattleWeathers []SeattleWeatherRequestDTO) {
	for _, value := range generated {
		var seattleWeather SeattleWeatherRequestDTO
		for i := 0; i < len(value.Values); i++ {
			switch i {
			case 1:
				seattleWeather.Date = value.Values[i].UserEnteredValue.StringValue
			case 2:
				seattleWeather.Precipitation = value.Values[i].UserEnteredValue.StringValue
			case 3:
				seattleWeather.TempMax = value.Values[i].UserEnteredValue.StringValue
			case 4:
				seattleWeather.TempMin = value.Values[i].UserEnteredValue.StringValue
			case 5:
				seattleWeather.Wind = value.Values[i].UserEnteredValue.StringValue
			case 6:
				seattleWeather.Weather = value.Values[i].UserEnteredValue.StringValue
			}
		}
		seattleWeathers = append(seattleWeathers, seattleWeather)
	}

	return
}

func FormatLocalSeattleWeather(data []map[string]interface{}) (seattleWeathers []SeattleWeatherRequestDTO) {
	for _, m := range data {
		var seattleWeather SeattleWeatherRequestDTO
		for k, v := range m {
			switch k {
			case "date":
				seattleWeather.Date = v.(string)
			case "precipitation":
				seattleWeather.Precipitation = v.(string)
			case "temp_max":
				seattleWeather.TempMax = v.(string)
			case "temp_min":
				seattleWeather.TempMin = v.(string)
			case "wind":
				seattleWeather.Wind = v.(string)
			case "weather":
				seattleWeather.Weather = v.(string)
			}
		}
		seattleWeathers = append(seattleWeathers, seattleWeather)
	}
	return
}
