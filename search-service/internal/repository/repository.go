package repository

import esv7 "github.com/elastic/go-elasticsearch/v7"

type Repositories struct {
	TaskRepository
}

func NewRepositories(client *esv7.Client) *Repositories {
	return &Repositories{
		TaskRepository: NewTask(client),
	}
}
