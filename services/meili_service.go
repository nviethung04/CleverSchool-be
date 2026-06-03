package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"be-lms/config"
)

type MeiliService interface {
	IndexDocuments(index string, docs []map[string]interface{}) error
	DeleteDocument(index string, id interface{}) error
}

type meiliService struct{}

func NewMeiliService() MeiliService { return &meiliService{} }

func (s *meiliService) buildHost(cfg config.Config) (string, error) {
	host := strings.TrimSpace(cfg.MeiliHost)
	if host == "" {
		return "", fmt.Errorf("MEILI_HOST empty")
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return strings.TrimRight(host, "/"), nil
}

func (s *meiliService) IndexDocuments(index string, docs []map[string]interface{}) error {
	cfg := config.LoadConfig()
	if !cfg.MeiliEnabled || len(docs) == 0 {
		return nil
	}
	host, err := s.buildHost(cfg)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(docs)
	url := host + "/indexes/" + index + "/documents"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if cfg.MeiliAPIKey != "" {
		req.Header.Set("X-Meili-API-Key", cfg.MeiliAPIKey)
		req.Header.Set("Authorization", "Bearer "+cfg.MeiliAPIKey)
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("meili index http status %d", resp.StatusCode)
	}
	return nil
}

func (s *meiliService) DeleteDocument(index string, id interface{}) error {
	cfg := config.LoadConfig()
	if !cfg.MeiliEnabled || id == nil {
		return nil
	}
	host, err := s.buildHost(cfg)
	if err != nil {
		return err
	}
	url := host + "/indexes/" + index + "/documents/" + fmt.Sprint(id)
	req, _ := http.NewRequest("DELETE", url, nil)
	if cfg.MeiliAPIKey != "" {
		req.Header.Set("X-Meili-API-Key", cfg.MeiliAPIKey)
		req.Header.Set("Authorization", "Bearer "+cfg.MeiliAPIKey)
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("meili delete http status %d", resp.StatusCode)
	}
	return nil
}
