package dto

type PresignRequest struct {
	Filename      string `json:"filename" binding:"required"`
	ContentType   string `json:"content_type,omitempty"`
	ExpiresSecond int64  `json:"expires_seconds,omitempty"`
	Folder        string `json:"folder,omitempty"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectURL string `json:"object_url"`
	Key       string `json:"key"`
	ExpiresIn int64  `json:"expires_in"`
}

type UploadCompleteRequest struct {
	Filename    string `json:"filename" binding:"required"`
	Key         string `json:"key" binding:"required"`
	URL         string `json:"url" binding:"required"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
}

type CompleteResponse struct {
	Message    string `json:"message"`
	FinalUrl   string `json:"final_url"`
}

type ExtractRequest struct {
	Filename    string `json:"filename" binding:"required"`
	Key         string `json:"key" binding:"required"`
	URL         string `json:"url" binding:"required"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	IsPowerPoint bool   `json:"is_powerpoint,omitempty"`
}

type ExtractResponse struct {
	Message    string `json:"message"`
	Folder     string `json:"folder"`
	Files      int32  `json:"files"`
	BaseKey    string `json:"base_key"`
	FinalUrl   string `json:"final_url"`
}
