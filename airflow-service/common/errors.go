package common

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrAirflowRequest = errors.New("airflow api request failed")
)
