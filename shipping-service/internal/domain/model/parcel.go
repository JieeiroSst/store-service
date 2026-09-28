package model

import (
	"errors"
	"fmt"
	"strings"
)

type Item struct {
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`
	Quantity   int    `json:"quantity"`
	WeightGram int    `json:"weight_gram"`
	Price      int64  `json:"price"`
}

type Parcel struct {
	WeightGram int    `json:"weight_gram"`
	LengthCm   int    `json:"length_cm"`
	WidthCm    int    `json:"width_cm"`
	HeightCm   int    `json:"height_cm"`
	Content    string `json:"content"`
	Items      []Item `json:"items"`
}

func (p Parcel) Validate() error {
	if p.WeightGram <= 0 || p.LengthCm <= 0 || p.WidthCm <= 0 || p.HeightCm <= 0 {
		return errors.New("parcel weight_gram, length_cm, width_cm and height_cm must be greater than zero")
	}
	if len(p.Items) == 0 {
		return errors.New("parcel needs at least one item")
	}
	for i, it := range p.Items {
		if strings.TrimSpace(it.Name) == "" || it.Quantity <= 0 {
			return fmt.Errorf("item %d needs a name and a quantity greater than zero", i+1)
		}
		if it.WeightGram < 0 || it.Price < 0 {
			return fmt.Errorf("item %d weight and price must not be negative", i+1)
		}
	}
	return nil
}
