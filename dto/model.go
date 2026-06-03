package dto

import "be-lms/prot"

type MyUploadResult struct {
	Data *prot.File
	Err  error
}

type MyExportResult struct {
	Url string
	Err error
}
