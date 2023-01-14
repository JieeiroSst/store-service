package repository

import "go.mongodb.org/mongo-driver/mongo"

type Repositories struct {
	Uploads
}

func NewRepositories(client *mongo.Client) *Repositories {
	return &Repositories{
		Uploads: NewUploadRepo(client),
	}
}
