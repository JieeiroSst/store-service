package dto

import "time"

type OfferRangeproductfileupload struct {
	ID               int       `json:"id"`
	FiledPath        string    `json:"filed_path"`
	Size             int       `json:"size"`
	DateUploaded     time.Time `json:"date_uploaded"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message"`
	DateProcessed    time.Time `json:"date_processed"`
	NumNewSkus       int       `json:"num_new_skus"`
	NumUnknowSkus    int       `json:"num_unknow_skus"`
	NumDuplicateSkus int       `json:"num_duplicate_skus"`
	RangeID          int       `json:"range_id"`
	UploadedById     int       `json:"uploaded_by_id"`
}
