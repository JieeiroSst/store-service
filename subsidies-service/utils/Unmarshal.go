package utils

import (
	"encoding/json"

	"github.com/JIeerioSst/subsidies-service/dto"
)

func Unmarshal(data []byte, discounts []dto.Discount) {
	json.Unmarshal(data, &discounts)
}
