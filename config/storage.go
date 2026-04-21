package config

import (
	"os"
)

type DiskConfig struct {
	Driver     string
	Root       string
	URL        string
	Visibility string
	Key        string
	Secret     string
	Region     string
	Bucket     string
	Endpoint   string
	PathStyle  bool
}

const (
	S3         = "s3"
	Public     = "public"
	PowerPoint = "power_point"
	Scorm      = "scorm"
	H5P        = "h5p"
)

var Disks map[string]DiskConfig

func InitDisks() {
	Disks = map[string]DiskConfig{
		Public: {
			Driver:     "local",
			Root:       "public",
			URL:        os.Getenv("APP_FULL_URL"),
			Visibility: "public",
		},
		S3: {
			Driver:    "s3",
			Key:       os.Getenv("AWS_ACCESS_KEY_ID"),
			Secret:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
			Region:    os.Getenv("AWS_DEFAULT_REGION"),
			Bucket:    os.Getenv("AWS_BUCKET"),
			URL:       os.Getenv("AWS_URL"),
			Endpoint:  os.Getenv("AWS_ENDPOINT"),
			PathStyle: os.Getenv("AWS_USE_PATH_STYLE_ENDPOINT") == "true",
		},
		PowerPoint: {
			Driver:     "s3",
			Root:       "power_point",
			URL:        os.Getenv("AWS_URL"),
			Visibility: "public",
			Key:        os.Getenv("AWS_ACCESS_KEY_ID"),
			Secret:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
			Region:     os.Getenv("AWS_DEFAULT_REGION"),
			Bucket:     os.Getenv("AWS_BUCKET"),
			Endpoint:   os.Getenv("AWS_ENDPOINT"),
			PathStyle:  os.Getenv("AWS_USE_PATH_STYLE_ENDPOINT") == "true",
		},
		Scorm: {
			Driver:     "scorm",
			Root:       "scorm",
			URL:        os.Getenv("APP_FULL_URL"),
			Visibility: "public",
		},
		H5P: {
			Driver:     "h5p",
			Root:       "h5p",
			URL:        os.Getenv("APP_FULL_URL"),
			Visibility: "public",
		},
	}
}
