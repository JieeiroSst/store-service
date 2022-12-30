package model

import "time"

type Media struct {
	Id         string
	FileName   string
	URL        string
	CreateDate time.Time
	UpdateDate time.Time
}

type UpdateMedia struct {
	FileName   string
	URL        string
	UpdateDate time.Time
}

type CreateMedia struct {
	Id         string
	FileName   string
	URL        string
	CreateDate time.Time
}
