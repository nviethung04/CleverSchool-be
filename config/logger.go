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
	"sync"
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

	// Async S3 upload
	logChan     chan []byte
	buffer      map[string]*bytes.Buffer // key -> buffer
	bufferMu    sync.Mutex
	flushTicker *time.Ticker
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func (w *S3Writer) Write(p []byte) (n int, err error) {
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

	// Production: Push vào channel để xử lý async, không block
	if w.logChan != nil {
		// Copy data để tránh race condition
		data := make([]byte, len(p))
		copy(data, p)

		select {
		case w.logChan <- data:
			// Successfully queued
		default:
			// Channel full, drop log to avoid blocking (có thể log warning nếu cần)
		}
	}

	return len(p), nil
}

// StartAsyncWorker khởi động goroutine để xử lý log async
func (w *S3Writer) StartAsyncWorker() {
	if w.Domain == "localhost" || w.Domain == "127.0.0.1" {
		return // Không cần async cho local
	}

	// Buffer channel với capacity 1000 entries
	w.logChan = make(chan []byte, 1000)
	w.buffer = make(map[string]*bytes.Buffer)
	w.stopChan = make(chan struct{})

	// Flush mỗi 10 giây
	w.flushTicker = time.NewTicker(10 * time.Second)

	w.wg.Add(1)
	go w.worker()
}

// Stop dừng worker và flush tất cả log còn lại
func (w *S3Writer) Stop() {
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}
	if w.stopChan != nil {
		close(w.stopChan)
	}
	w.wg.Wait()

	// Flush lần cuối
	w.flushAll()
}

// worker xử lý log entries từ channel
func (w *S3Writer) worker() {
	defer w.wg.Done()

	for {
		select {
		case data, ok := <-w.logChan:
			if !ok {
				return // Channel closed
			}
			w.appendToBuffer(data)

		case <-w.flushTicker.C:
			w.flushAll()

		case <-w.stopChan:
			return
		}
	}
}

// appendToBuffer thêm log vào buffer theo key (ngày)
func (w *S3Writer) appendToBuffer(data []byte) {
	key := w.getS3Key()

	w.bufferMu.Lock()
	defer w.bufferMu.Unlock()

	if w.buffer[key] == nil {
		w.buffer[key] = &bytes.Buffer{}
	}
	w.buffer[key].Write(data)
}

// flushAll upload tất cả buffer lên S3
func (w *S3Writer) flushAll() {
	w.bufferMu.Lock()
	buffers := make(map[string]*bytes.Buffer)
	for k, v := range w.buffer {
		buffers[k] = v
		w.buffer[k] = &bytes.Buffer{} // Reset buffer
	}
	w.bufferMu.Unlock()

	for key, buf := range buffers {
		if buf.Len() == 0 {
			continue
		}
		w.uploadToS3(key, buf.Bytes())
	}
}

// uploadToS3 upload log lên S3 (đồng bộ, nhưng chạy trong goroutine riêng)
func (w *S3Writer) uploadToS3(key string, newData []byte) {
	if w.Client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Đọc nội dung cũ (nếu có)
	var oldData []byte
	obj, err := w.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(w.Bucket),
		Key:    aws.String(key),
	})
	if err == nil && obj.Body != nil {
		oldData, _ = io.ReadAll(obj.Body)
		obj.Body.Close()
	}

	// Append log mới
	finalData := append(oldData, newData...)

	// Ghi đè lại object
	_, err = w.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(w.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(finalData),
	})
	if err != nil {
		// Log error nhưng không block (có thể ghi vào stderr)
		fmt.Fprintf(os.Stderr, "Failed to upload log to S3: %v\n", err)
	}
}

// getS3Key trả về S3 key cho ngày hiện tại
func (w *S3Writer) getS3Key() string {
	return fmt.Sprintf("%s/%s-%s.log",
		w.KeyPrefix,
		w.Domain,
		time.Now().Format("2006-01-02"),
	)
}

var Log *logrus.Logger
var s3WriterInstance *S3Writer // Global instance để có thể stop khi shutdown

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

	var s3Client *s3.Client
	if domain != "localhost" && domain != "127.0.0.1" {
		// Load AWS config nếu không phải local
		awsCfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic(err)
		}
		s3Client = s3.NewFromConfig(awsCfg)
	}

	// Custom writer
	s3Writer := &S3Writer{
		Client:    s3Client,
		Bucket:    os.Getenv("AWS_BUCKET"),
		KeyPrefix: "logs",
		Domain:    domain,
	}

	// Khởi động async worker cho S3 (chỉ khi không phải local)
	s3Writer.StartAsyncWorker()
	s3WriterInstance = s3Writer // Lưu để có thể stop khi shutdown

	Log = logrus.New()

	// Ghi log ra stdout + writer (S3 hoặc local file)
	Log.SetOutput(io.MultiWriter(os.Stdout, s3Writer))

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	})

	Log.SetLevel(logrus.InfoLevel)

	if domain == "localhost" || domain == "127.0.0.1" {
		Log.Infof("📝 Logger initialized, writing to ./logs folder (local mode)")
	} else {
		Log.Infof("🚀 Logger initialized, writing to S3 bucket: %s", os.Getenv("AWS_BUCKET"))

		// Thêm hook Discord nếu có webhook
		discordWebhook := LoadConfig().DiscordHookUrl
		if discordWebhook != "" {
			Log.AddHook(NewAsyncDiscordHook(discordWebhook))
		}
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

// StopLogger dừng async S3 logger worker và flush tất cả log còn lại
// Nên gọi hàm này khi app shutdown để đảm bảo không mất log
func StopLogger() {
	if s3WriterInstance != nil {
		s3WriterInstance.Stop()
	}
}
