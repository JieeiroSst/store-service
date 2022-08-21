package minio

import "mime/multipart"

type UploadFileArgs struct {
	UserMetaData map[string]string
	FileHeader   *multipart.FileHeader
	FileName     string
}

type UploadObjectArgs struct {
	FileHeader *multipart.FileHeader
}

type UploadObjectResponse struct {
	URL      string
	FileName string
}