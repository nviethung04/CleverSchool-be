package services

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/datatypes"
)

type ContentUserDataInput struct {
	ContentId    string                 `json:"contentId"`
	ContextId    string                 `json:"contextId"`
	DataType     string                 `json:"dataType"`
	Invalidate   bool                   `json:"invalidate"`
	Preload      bool                   `json:"preload"`
	SubContentId string                 `json:"subContentId"`
	UserState    map[string]interface{} `json:"userState"`
	UserId       string                 `json:"userId"`
}

type Content struct {
	ContentID  int64                  `json:"contentId"`
	Title      string                 `json:"title"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type ContentScore struct {
	ContentId string `json:"contentId"`
	Score     string `json:"score"`
	MaxScore  string `json:"maxScore"`
	Opened    string `json:"opened"`
	Finished  string `json:"finished"`
	Time      string `json:"time"`
	UserId    string `json:"userId"`
}

type H5PUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Content      map[string]interface{} `json:"content"`
		ContentTypes map[string]interface{} `json:"contentTypes"`
		H5P          struct {
			Title                 string   `json:"title"`
			Language              string   `json:"language"`
			MainLibrary           string   `json:"mainLibrary"`
			EmbedTypes            []string `json:"embedTypes"`
			License               string   `json:"license"`
			DefaultLanguage       string   `json:"defaultLanguage"`
			PreloadedDependencies []struct {
				MachineName  string `json:"machineName"`
				MajorVersion int    `json:"majorVersion"`
				MinorVersion int    `json:"minorVersion"`
			} `json:"preloadedDependencies"`
		} `json:"h5p"`
	} `json:"data"`
}

type StringInt string

func (s *StringInt) UnmarshalJSON(b []byte) error {
	// thử parse int
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*s = StringInt(strconv.Itoa(i))
		return nil
	}
	// fallback string
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = StringInt(str)
		return nil
	}
	return fmt.Errorf("invalid value for StringInt: %s", string(b))
}

type UploadResponse struct {
	ContentId StringInt              `json:"contentId"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type H5pService interface {
	Upload(c *gin.Context) (*prot.File, error)
	ListContent(c *gin.Context) ([]models.H5pContent, int64, error)
	ShowContent(c *gin.Context, id string) (*models.H5pContent, error)
	UpdateContent(c *gin.Context, id string) (*prot.H5P, error)
	DeleteContent(c *gin.Context, id string) error
	StoreScore(c *gin.Context) error
	StoreContentUserData(c *gin.Context) error
	GetContentUserData(c *gin.Context) map[string]interface{}
	GetContentUserDataByContentIdAndUser(c *gin.Context) []map[string]interface{}
}

type h5pService struct {
	repo repositories.H5pRepository
}

func NewH5pService(repo repositories.H5pRepository) H5pService {
	return &h5pService{repo}
}

func (s *h5pService) Upload(c *gin.Context) (*prot.File, error) {
	h5pUrl := config.LoadConfig().H5PUrl
	config.Log.Info("Uploading h5p: ", h5pUrl)
	token := c.GetHeader("Token")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get h5p file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("h5p", header.Filename)
	if err != nil {
		config.Log.Error("failed to create form file: ", err)
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		config.Log.Error("failed to copy file: ", err)
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		config.Log.Error("failed to close writer: ", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", h5pUrl+"/h5p/ajax?action=library-upload", body)
	if err != nil {
		config.Log.Error("failed to create request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "h5p_token="+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error("failed to upload h5p: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		config.Log.Error("failed to read response body: ", err)
		return nil, err
	}

	var uploadResp H5PUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		config.Log.Error("invalid upload response: ", err)
		return nil, fmt.Errorf("invalid upload response: %w", err)
	}
	if !uploadResp.Success {
		config.Log.Error("upload failed: ", uploadResp.Message)
		return nil, fmt.Errorf("upload failed: %s", uploadResp.Message)
	}

	lastDep := uploadResp.Data.H5P.PreloadedDependencies[len(uploadResp.Data.H5P.PreloadedDependencies)-1]
	library := fmt.Sprintf("%s %d.%d",
		uploadResp.Data.H5P.MainLibrary,
		lastDep.MajorVersion,
		lastDep.MinorVersion,
	)

	payload := map[string]interface{}{
		"library": library,
		"params": map[string]interface{}{
			"params": uploadResp.Data.Content,
			"metadata": map[string]interface{}{
				"title":   uploadResp.Data.H5P.Title,
				"license": uploadResp.Data.H5P.License,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		config.Log.Error("failed to marshal payload: ", err)
		return nil, err
	}

	// Push content
	config.Log.Info("Pushing content: ", h5pUrl+"/h5p")
	req2, err := http.NewRequest("POST", h5pUrl+"/h5p", bytes.NewBuffer(payloadBytes))
	if err != nil {
		config.Log.Error("failed to create request: ", err)
		return nil, err
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", "h5p_token="+token)

	resp2, err := client.Do(req2)
	if err != nil {
		config.Log.Error("failed to push content: ", err)
		return nil, err
	}
	defer resp2.Body.Close()

	respBody2, err := io.ReadAll(resp2.Body)
	if err != nil {
		config.Log.Error("failed to read response body: ", err)
		return nil, err
	}

	var finalResp UploadResponse
	if err := json.Unmarshal(respBody2, &finalResp); err != nil {
		config.Log.Error("invalid final response: ", err)
		return nil, fmt.Errorf("invalid final response: %w", err)
	}

	diskName := config.Public
	rootPath := "h5p"
	parentZero := int64(0)
	parentId := int64(0)

	mediaRepo := repositories.NewMediaRepository()

	parentRootFolder, err := mediaRepo.FindFolderByFilePathAndParent("", int64(0), diskName)

	if err == nil && parentRootFolder.ID != 0 {
		parentZero = parentRootFolder.ID
	}

	rootFolder, err := mediaRepo.FindFolderByFilePathAndParent(rootPath, parentZero, diskName)

	if err == nil && rootFolder.ID != 0 {
		parentId = rootFolder.ID
	}

	staticUrl := string(finalResp.ContentId)
	filePath := string(finalResp.ContentId)

	media := &models.Media{
		FolderID:      parentId,
		FileName:      finalResp.Metadata["title"].(string),
		FilePath:      filePath,
		FileType:      utils.StringPtr("h5p"),
		FileExtension: utils.StringPtr("h5p"),
		DiskName:      utils.StringPtr(config.H5P),
		StaticURL:     &staticUrl,
		Type:          "file",
	}

	if err := mediaRepo.Save(media); err != nil {
		config.Log.Error("failed to save media: ", err)
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	h5pRepo := repositories.NewH5pContentRepository()

	dataMap := uploadResp.Data.Content
	jsonBytes, err := json.Marshal(dataMap)
	if err != nil {
		return nil, err
	}

	metaDataMap := map[string]interface{}{
		"title":   uploadResp.Data.H5P.Title,
		"license": uploadResp.Data.H5P.License,
	}

	metaDataBytes, err := json.Marshal(metaDataMap)
	if err != nil {
		return nil, err
	}

	h5p := &models.H5pContent{
		ContentID:  string(finalResp.ContentId),
		Title:      uploadResp.Data.H5P.Title,
		Library:    library,
		Parameters: jsonBytes,
		Metadata:   metaDataBytes,
	}

	if err := h5pRepo.CreateOrUpdate(h5p); err != nil {
		config.Log.Error("failed to save h5p: ", err)
		return nil, fmt.Errorf("failed to save h5p: %w", err)
	}

	return &prot.File{
		Id:            int32(media.ID),
		FileName:      media.FileName,
		Url:           staticUrl,
		FileSize:      strconv.FormatInt(utils.DerefInt64(media.FileSize), 10),
		FileMimeType:  utils.DerefStr(media.FileType),
		FileExtension: utils.DerefStr(media.FileExtension),
		DiskName:      utils.DerefStr(media.DiskName),
		Path:          filePath,
		Type:          media.Type,
	}, nil
}

func (s *h5pService) ListContent(c *gin.Context) ([]models.H5pContent, int64, error) {
	allowedFilters := []string{}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	contents, rows, err := s.repo.FindAll()

	if err != nil {
		return nil, 0, err
	}

	return contents, rows, err
}

func (s *h5pService) ShowContent(c *gin.Context, id string) (*models.H5pContent, error) {
	content, err := s.repo.FindByContentID(id)

	if err != nil {
		return nil, err
	}

	return content, err
}

func (s *h5pService) UpdateContent(c *gin.Context, id string) (*prot.H5P, error) {
	req, err, _ := utils.GetBody[*prot.H5P](c, func() *prot.H5P {
		return &prot.H5P{}
	})
	if err != nil {
		return nil, err
	}

	content := models.H5pContent{
		ContentID:  req.ContentId,
		Title:      req.Title,
		Parameters: structToJSON(req.Parameters),
		Metadata:   structToJSON(req.Metadata),
	}

	s.repo.SetContext(c)
	err = s.repo.UpdateByContentId(&content)

	if err != nil {
		return nil, err
	}

	return req, err
}

func (s *h5pService) DeleteContent(c *gin.Context, id string) error {
	h5pUrl := config.LoadConfig().H5PUrl
	h5pUrl = h5pUrl + "/h5p/" + id
	token := c.GetHeader("Token")
	config.Log.Info("Url: ", h5pUrl)

	// Gọi DELETE tới H5P server
	req, err := http.NewRequest("DELETE", h5pUrl, nil)
	if err != nil {
		config.Log.Error("failed to create DELETE request: ", err)
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "h5p_token="+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error("failed to call DELETE /h5p: ", err)
		return fmt.Errorf("failed to call DELETE /h5p: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		config.Log.Error("h5p server returned error: ", string(body))
		return fmt.Errorf("h5p server returned error: %s", string(body))
	}

	// Xóa trong DB sau khi gọi H5P thành công
	s.repo.SetContext(c)
	if err := s.repo.DeleteByContentId(id); err != nil {
		return fmt.Errorf("failed to delete content in DB: %w", err)
	}

	return nil
}

func (s *h5pService) StoreScore(c *gin.Context) error {
	var input ContentScore
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	openedTime, err := safeParseUnixTimestamp(input.Opened)
	if err != nil {
		return err
	}

	finishedTime, err := safeParseUnixTimestamp(input.Finished)
	if err != nil {
		return err
	}

	timeVal, err := safeParseTimestamp(input.Time)
	if err != nil {
		timeVal = nil
	}

	scoreStr := input.Score
	maxScoreStr := input.MaxScore

	maxScoreFloat64, err := strconv.ParseFloat(maxScoreStr, 32)
	if err != nil {
		maxScoreFloat64 = 0
	}

	scoreFloat64, err := strconv.ParseFloat(scoreStr, 32)
	if err != nil {
		scoreFloat64 = 0
	}

	userId := utils.GetCurrentUserId(c)

	record := models.H5pContentScore{
		ContentId: input.ContentId,
		MaxScore:  float32(maxScoreFloat64),
		Score:     float32(scoreFloat64),
		Opened:    openedTime,
		Finished:  finishedTime,
		Time:      timeVal,
		UserId:    int64(userId),
	}

	err = s.repo.StoreScore(record)

	if err != nil {
		return err
	}

	return nil
}

func (s *h5pService) StoreContentUserData(c *gin.Context) error {
	var input ContentUserDataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}

	userStateBytes, err := json.Marshal(input.UserState)
	if err != nil {
		return err
	}

	userIdRaw, _ := c.Get("userID")

	var userIdStr string
	switch v := userIdRaw.(type) {
	case string:
		userIdStr = v
	case int:
		userIdStr = strconv.Itoa(v)
	case int64:
		userIdStr = strconv.FormatInt(v, 10)
	case float64:
		userIdStr = strconv.FormatInt(int64(v), 10)
	default:
		userIdStr = input.UserId
	}

	record := models.H5pContentUserData{
		ContentId:    input.ContentId,
		ContextId:    input.ContextId,
		DataType:     input.DataType,
		Invalidate:   input.Invalidate,
		Preload:      input.Preload,
		SubContentId: input.SubContentId,
		UserState:    string(userStateBytes),
		UserId:       userIdStr,
	}

	err = s.repo.StoreContentUserData(record)
	if err != nil {
		return err
	}

	return nil
}

func (s *h5pService) GetContentUserData(c *gin.Context) map[string]interface{} {
	contentId := c.Param("contentId")
	userId := c.Param("userId")

	rawResult, err := s.repo.GetContentUserData(contentId, userId)

	if err != nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"user_state": rawResult.UserState,
		"invalidate": rawResult.Invalidate,
		"preload":    rawResult.Preload,
	}
}

func (s *h5pService) GetContentUserDataByContentIdAndUser(c *gin.Context) []map[string]interface{} {
	contentId := c.Param("contentId")
	userId := c.Param("userId")

	rawResults, err := s.repo.GetContentUserDataByContentIdAndUser(contentId, userId)

	if err != nil {
		return []map[string]interface{}{}
	}

	var result []map[string]interface{}
	for _, r := range rawResults {
		result = append(result, map[string]interface{}{
			"state": r.UserState,
		})
	}

	return result
}

func structToJSON(s *structpb.Struct) datatypes.JSON {
	if s == nil {
		return nil
	}
	m := s.AsMap() // *structpb.Struct -> map[string]interface{}
	b, _ := json.Marshal(m)
	return b
}

func safeParseTimestamp(ts string) (*time.Time, error) {
	if ts == "" || ts == "0000-00-00 00:00:00" {
		return nil, nil
	}
	t, err := parseUnixTimestamp(ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func safeParseUnixTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, nil
	}
	return parseUnixTimestamp(ts)
}

func parseUnixTimestamp(ts string) (time.Time, error) {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

func convertToDatatypesJSON(m map[string]interface{}) (datatypes.JSON, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

