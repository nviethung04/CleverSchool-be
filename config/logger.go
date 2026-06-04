package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

type S3Writer struct {
	Client    *s3.Client
	Bucket    string
	KeyPrefix string
	Domain    string
}

func (w *S3Writer) Write(p []byte) (n int, err error) {
	if w.Client == nil || w.Bucket == "" {
		return len(p), nil
	}

	// Nếu chạy local → ghi file trong local, không push S3
	if w.Domain == "localhost" || w.Domain == "127.0.0.1" {
		logDir := "./logs"
		_ = os.MkdirAll(logDir, 0755)

		fileName := fmt.Sprintf("%s-%s.log",
			w.Domain,
			time.Now().Format("2006-01-02"),
		)
		path := filepath.Join(logDir, fileName)

		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		_, err = f.Write(p)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}

	// Ngược lại → ghi log lên S3
	key := fmt.Sprintf("%s/%s-%s.log",
		w.KeyPrefix,
		w.Domain,
		time.Now().Format("2006-01-02"),
	)

	// Đọc nội dung cũ (nếu có)
	var oldData []byte
	obj, err := w.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(w.Bucket),
		Key:    aws.String(key),
	})
	if err == nil && obj.Body != nil {
		defer obj.Body.Close()
		oldData, _ = io.ReadAll(obj.Body)
	}

	// Append log mới
	newData := append(oldData, p...)

	// Ghi đè lại object
	_, err = w.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(w.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(newData),
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var Log *logrus.Logger

func InitLogger() {
	// set timezone VN
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	// Domain để đặt tên log
	domain := os.Getenv("API_DOMAIN")
	if domain == "" {
		domain = "localhost"
	}
	domain = ExtractDomainName(domain)

	Log = logrus.New()

	writers := []io.Writer{os.Stdout}
	if S3Configured() {
		awsCfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic(err)
		}
		writers = append(writers, &S3Writer{
			Client:    s3.NewFromConfig(awsCfg),
			Bucket:    os.Getenv("AWS_BUCKET"),
			KeyPrefix: "logs",
			Domain:    domain,
		})
	}

	Log.SetOutput(io.MultiWriter(writers...))

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	})

	Log.SetLevel(logrus.InfoLevel)

	if S3Configured() {
		Log.Infof("Logger initialized (stdout + S3 bucket: %s)", os.Getenv("AWS_BUCKET"))
	} else if domain == "localhost" || domain == "127.0.0.1" {
		Log.Info("Logger initialized (stdout only; local — no S3)")
	} else {
		Log.Info("Logger initialized (stdout only; S3 not configured)")
	}

	discordWebhook := LoadConfig().DiscordHookUrl
	if discordWebhook != "" {
		Log.AddHook(NewAsyncDiscordHook(discordWebhook))
	}
}

func ExtractDomainName(rawURL string) string {
	u, err := url.Parse(rawURL)
	var host string
	if err != nil || u.Hostname() == "" {
		host = strings.TrimPrefix(rawURL, "http://")
		host = strings.TrimPrefix(host, "https://")
		host = strings.Split(host, "/")[0]
		host = strings.Split(host, ":")[0]
	} else {
		host = u.Hostname()
	}

	// Nếu là localhost hoặc 127.0.0.1 thì giữ nguyên
	if host == "localhost" || host == "127.0.0.1" {
		return host
	}

	if idx := strings.Index(host, "-api"); idx != -1 {
		host = host[:idx]
	} else {
		if dot := strings.Index(host, "."); dot != -1 {
			host = host[:dot]
		}
	}

	return strings.ReplaceAll(host, "-", "_")
}
