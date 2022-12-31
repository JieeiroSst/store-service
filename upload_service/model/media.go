package model

import "time"

type Media struct {
	Id         string    `json:"id" bson:"_id,omitempty"`
	FileName   string    `json:"file_name" bson:"file_name,omitempty"`
	URL        string    `json:"url" bson:"url,omitempty"`
	CreateDate time.Time `json:"create_date" bson:"create_date,omitempty"`
	UpdateDate time.Time `json:"update_date" bson:"update_date"`
}

type UpdateMedia struct {
	FileName   string    `json:"file_name" bson:"file_name,omitempty"`
	URL        string    `json:"url" bson:"url,omitempty"`
	UpdateDate time.Time `json:"update_date" bson:"update_date,omitempty"`
}

type CreateMedia struct {
	Id         string    `json:"id" bson:"_id,omitempty"`
	FileName   string    `json:"file_name" bson:"file_name,omitempty"`
	URL        string    `json:"url" bson:"url,omitempty"`
	CreateDate time.Time `json:"create_date" bson:"create_date,omitempty"`
}
