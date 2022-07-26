package minio

import "mime/multipart"

type UploadFileArgs struct {
	UserMetaData map[string]string
	File         multipart.File
	FileHeader   *multipart.FileHeader
	FileName     string
}

type UploadObjectArgs struct {
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type UploadObjectResponse struct {
	URL      string
	FileName string
}