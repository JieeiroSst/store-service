package repository

import "go.mongodb.org/mongo-driver/mongo"

type Repositories struct {
	Uploads
}

func NewRepositories(collection *mongo.Collection) *Repositories {
	return &Repositories{
		Uploads: NewUploadRepo(collection),
	}
}
