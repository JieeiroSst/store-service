package repository

import "go.mongodb.org/mongo-driver/mongo"

type Repositories struct {
	Messages
}

func NewRepositories(db *mongo.Client) *Repositories {
	return &Repositories{
		Messages: NewMessageRepo(db),
	}
}
