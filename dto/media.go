package dto

type MediaDTO struct {
	ID       int64  `json:"id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FullPath string `json:"full_path"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	URL      string `json:"url"`
}

type PowerPoint struct {
	FolderName string `json:"folder_name"`
	FileSize   int64  `json:"file_size"`
	Filename   string `json:"filename"`
	CreatedAt  string `json:"created_at"`
	StaticURL  string `json:"static_url"`
}

type PowerPointInfo struct {
	FolderName string `json:"folder_name"`
	FileSize   int64  `json:"file_size"`
	CreatedAt  string `json:"created_at"`
	StaticURL  string `json:"static_url"`
}

type Scorm struct {
	FolderName string `json:"folder_name"`
	FileSize   int64  `json:"file_size"`
	Filename   string `json:"filename"`
	CreatedAt  string `json:"created_at"`
	StaticURL  string `json:"static_url"`
}

type ScormInfo struct {
	FolderName string `json:"folder_name"`
	FileSize   int64  `json:"file_size"`
	CreatedAt  string `json:"created_at"`
	StaticURL  string `json:"static_url"`
}
