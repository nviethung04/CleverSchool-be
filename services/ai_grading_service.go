package services

import (
	"be-lms/config"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIGradingService interface {
	GradeWritingEssay(essayText string, rubric string) (*WritingGradeResult, error)
	GradeSpeakingAudio(audioURL string, rubric string) (*SpeakingGradeResult, error)
}

type aiGradingService struct {
	openAIKey string
	geminiKey string
	useGemini bool // Flag để chọn dùng Gemini hay OpenAI
}

func NewAIGradingService() AIGradingService {
	geminiKey := os.Getenv("GEMINI_API_KEY")
	openAIKey := os.Getenv("OPENAI_API_KEY")

	// Ưu tiên dùng Gemini nếu có key (vì free)
	useGemini := geminiKey != ""

	return &aiGradingService{
		openAIKey: openAIKey,
		geminiKey: geminiKey,
		useGemini: useGemini,
	}
}

// WritingGradeResult - Kết quả chấm writing
type WritingGradeResult struct {
	OverallScore      float32  `json:"overall_score"`      // Điểm tổng (0-10)
	GrammarScore      float32  `json:"grammar_score"`      // Điểm ngữ pháp
	VocabularyScore   float32  `json:"vocabulary_score"`   // Điểm từ vựng
	OrganizationScore float32  `json:"organization_score"` // Điểm tổ chức bài
	ContentScore      float32  `json:"content_score"`      // Điểm nội dung
	Feedback          string   `json:"feedback"`           // Nhận xét chi tiết
	Strengths         []string `json:"strengths"`          // Điểm mạnh
	Improvements      []string `json:"improvements"`       // Cần cải thiện
}

// SpeakingGradeResult - Kết quả chấm speaking
type SpeakingGradeResult struct {
	OverallScore       float32  `json:"overall_score"`       // Điểm tổng (0-10)
	PronunciationScore float32  `json:"pronunciation_score"` // Điểm phát âm
	FluencyScore       float32  `json:"fluency_score"`       // Điểm độ trưng
	GrammarScore       float32  `json:"grammar_score"`       // Điểm ngữ pháp
	VocabularyScore    float32  `json:"vocabulary_score"`    // Điểm từ vựng
	Transcript         string   `json:"transcript"`          // Bản transcript
	Feedback           string   `json:"feedback"`            // Nhận xét chi tiết
	Strengths          []string `json:"strengths"`           // Điểm mạnh
	Improvements       []string `json:"improvements"`        // Cần cải thiện
}

// OpenAI API structures
type openAIRequest struct {
	Model       string      `json:"model"`
	Messages    []aiMessage `json:"messages"`
	Temperature float32     `json:"temperature,omitempty"`
}

type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Google Gemini API structures
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// GradeWritingEssay - Chấm bài writing
func (s *aiGradingService) GradeWritingEssay(essayText string, rubric string) (*WritingGradeResult, error) {
	// Check API keys
	if s.useGemini && s.geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
	}
	if !s.useGemini && s.openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	// Tạo prompt
	systemPrompt := `You are an expert English teacher specializing in essay grading. 
Score essays based on these criteria (0-10 scale):
1. Grammar & Mechanics (25%)
2. Vocabulary & Word Choice (25%)
3. Organization & Coherence (25%)
4. Content & Ideas (25%)

IMPORTANT: Provide ALL feedback text in Vietnamese language.

Provide detailed feedback in JSON format:
{
    "overall_score": 8.5,
    "grammar_score": 8.0,
    "vocabulary_score": 9.0,
    "organization_score": 8.5,
    "content_score": 9.0,
    "feedback": "Bài viết tổng thể rất xuất sắc...",
    "strengths": ["Sử dụng từ nối tốt", "Luận điểm rõ ràng"],
    "improvements": ["Cần thêm ví dụ", "Một số lỗi ngữ pháp"]
}`

	userPrompt := fmt.Sprintf("Rubric/Question: %s\n\nStudent Essay:\n%s\n\nPlease grade this essay and provide feedback in JSON format with Vietnamese feedback text.", rubric, essayText)

	// Gọi API tùy theo provider
	var result string
	var err error

	if s.useGemini {
		result, err = s.callGemini(systemPrompt, userPrompt)
	} else {
		result, err = s.callOpenAI(systemPrompt, userPrompt)
	}

	if err != nil {
		return nil, err
	}

	// Log raw response for debugging
	config.Log.Infof("📝 Raw AI response (first 200 chars): %s", truncateString(result, 200))

	// Clean response - remove markdown code blocks if present
	cleanedResult := cleanJSONResponse(result)
	config.Log.Infof("✨ Cleaned response (first 200 chars): %s", truncateString(cleanedResult, 200))

	// Parse JSON response
	var gradeResult WritingGradeResult
	if err := json.Unmarshal([]byte(cleanedResult), &gradeResult); err != nil {
		// Nếu không parse được JSON, trả về feedback text
		config.Log.Warnf("Could not parse JSON response: %v, returning plain feedback", err)
		return &WritingGradeResult{
			OverallScore: 0,
			Feedback:     result,
			Strengths:    []string{},
			Improvements: []string{},
		}, nil
	}

	// Validate và đảm bảo arrays không nil
	if gradeResult.Strengths == nil {
		gradeResult.Strengths = []string{}
	}
	if gradeResult.Improvements == nil {
		gradeResult.Improvements = []string{}
	}

	return &gradeResult, nil
}

// GradeSpeakingAudio - Chấm bài speaking (từ audio URL)
func (s *aiGradingService) GradeSpeakingAudio(audioURL string, rubric string) (*SpeakingGradeResult, error) {
	// Gemini 2.0 hỗ trợ audio input trực tiếp
	if s.useGemini && s.geminiKey != "" {
		return s.gradeSpeakingWithGemini(audioURL, rubric)
	}

	// OpenAI không hỗ trợ audio analysis trực tiếp, cần Whisper trước
	if s.openAIKey != "" {
		// Bước 1: Download audio
		audioData, err := s.downloadAudio(audioURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download audio: %w", err)
		}

		// Bước 2: Transcribe với Whisper
		transcript, err := s.transcribeWithWhisper(audioData)
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %w", err)
		}

		// Bước 3: Chấm điểm dựa trên transcript
		return s.gradeSpeakingFromTranscript(transcript, rubric)
	}

	return nil, fmt.Errorf("no AI service available for speaking grading")
}

// gradeSpeakingWithGemini - Chấm speaking trực tiếp từ audio URL bằng Gemini
func (s *aiGradingService) gradeSpeakingWithGemini(audioURL string, rubric string) (*SpeakingGradeResult, error) {
	// Download audio để upload lên Gemini
	audioData, err := s.downloadAudio(audioURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}

	systemPrompt := `You are an expert English speaking examiner. 
Analyze the audio and score speaking based on these criteria (0-10 scale):
1. Pronunciation & Clarity (25%)
2. Fluency & Coherence (25%)
3. Grammar & Accuracy (25%)
4. Vocabulary Range (25%)

IMPORTANT: Provide ALL feedback text in Vietnamese language.

Provide detailed feedback in JSON format:
{
    "overall_score": 8.5,
    "pronunciation_score": 8.0,
    "fluency_score": 9.0,
    "grammar_score": 8.5,
    "vocabulary_score": 9.0,
    "transcript": "The transcribed text...",
    "feedback": "Bài nói tổng thể rất xuất sắc...",
    "strengths": ["Phát âm rõ ràng", "Độ lưu loát tốt"],
    "improvements": ["Sử dụng từ vựng nâng cao hơn", "Một số lỗi ngữ pháp"]
}`

	userPrompt := fmt.Sprintf("Rubric/Question: %s\n\nPlease analyze the audio and provide speaking assessment in JSON format with Vietnamese feedback text.", rubric)

	// Gọi Gemini với audio
	result, err := s.callGeminiWithAudio(systemPrompt, userPrompt, audioData)
	if err != nil {
		return nil, err
	}

	// Log raw response for debugging
	config.Log.Infof("🎤 Raw Gemini audio response (first 200 chars): %s", truncateString(result, 200))

	// Clean JSON response (remove markdown code blocks if present)
	cleanedResult := cleanJSONResponse(result)
	config.Log.Infof("✨ Cleaned audio response (first 200 chars): %s", truncateString(cleanedResult, 200))

	// Parse JSON response
	var gradeResult SpeakingGradeResult
	if err := json.Unmarshal([]byte(cleanedResult), &gradeResult); err != nil {
		config.Log.Warnf("Could not parse JSON response, returning plain feedback")
		return &SpeakingGradeResult{
			OverallScore: 0,
			Feedback:     result,
		}, nil
	}

	// Ensure arrays are never null
	if gradeResult.Strengths == nil {
		gradeResult.Strengths = []string{}
	}
	if gradeResult.Improvements == nil {
		gradeResult.Improvements = []string{}
	}

	return &gradeResult, nil
}

// gradeSpeakingFromTranscript - Chấm speaking từ transcript text (dùng cho OpenAI)
func (s *aiGradingService) gradeSpeakingFromTranscript(transcript string, rubric string) (*SpeakingGradeResult, error) {
	systemPrompt := `You are an expert English speaking examiner. 
Score speaking based on the transcript (0-10 scale):
1. Grammar & Accuracy (33%)
2. Vocabulary Range (33%)
3. Coherence & Organization (34%)

Note: Cannot assess pronunciation from text transcript.

IMPORTANT: Provide ALL feedback text in Vietnamese language.

Provide detailed feedback in JSON format:
{
    "overall_score": 8.5,
    "pronunciation_score": 0,
    "fluency_score": 9.0,
    "grammar_score": 8.5,
    "vocabulary_score": 9.0,
    "transcript": "The transcribed text...",
    "feedback": "Bài nói tổng thể rất tốt...",
    "strengths": ["Từ vựng phong phú", "Ý tưởng rõ ràng"],
    "improvements": ["Một số lỗi ngữ pháp"]
}`

	userPrompt := fmt.Sprintf("Rubric/Question: %s\n\nStudent Speaking Transcript:\n%s\n\nPlease grade this speaking performance and provide feedback in JSON format with Vietnamese feedback text.", rubric, transcript)

	result, err := s.callOpenAI(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Clean JSON response (remove markdown code blocks if present)
	result = cleanJSONResponse(result)

	var gradeResult SpeakingGradeResult
	if err := json.Unmarshal([]byte(result), &gradeResult); err != nil {
		config.Log.Warnf("Could not parse JSON response")
		return &SpeakingGradeResult{
			OverallScore: 0,
			Transcript:   transcript,
			Feedback:     result,
		}, nil
	}

	gradeResult.Transcript = transcript

	// Ensure arrays are never null
	if gradeResult.Strengths == nil {
		gradeResult.Strengths = []string{}
	}
	if gradeResult.Improvements == nil {
		gradeResult.Improvements = []string{}
	}

	return &gradeResult, nil
}

// callGeminiWithAudio - Gọi Gemini API với audio input
func (s *aiGradingService) callGeminiWithAudio(systemPrompt, userPrompt string, audioData []byte) (string, error) {
	combinedPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": combinedPrompt},
					{
						"inline_data": map[string]string{
							"mime_type": "audio/mpeg",
							"data":      base64.StdEncoding.EncodeToString(audioData),
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", s.geminiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second} // Audio cần timeout cao hơn
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	config.Log.Infof("🎤 Gemini audio analysis successful")

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// downloadAudio - Tải audio từ URL
func (s *aiGradingService) downloadAudio(audioURL string) ([]byte, error) {
	resp, err := http.Get(audioURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download audio: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// transcribeAudio - Chuyển audio thành text bằng Whisper API
func (s *aiGradingService) transcribeAudio(audioData []byte) (string, error) {
	// Nếu có Gemini key, có thể dùng Gemini để transcribe (free)
	// Nếu không, dùng OpenAI Whisper

	if s.openAIKey != "" {
		return s.transcribeWithWhisper(audioData)
	}

	return "", fmt.Errorf("no transcription service available - need OPENAI_API_KEY for Whisper")
}

// transcribeWithWhisper - Sử dụng OpenAI Whisper API
func (s *aiGradingService) transcribeWithWhisper(audioData []byte) (string, error) {
	// Tạo multipart form data
	body := &bytes.Buffer{}

	// Tạo boundary cho multipart
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"

	// Write file part
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"audio.mp3\"\r\n")
	body.WriteString("Content-Type: audio/mpeg\r\n\r\n")
	body.Write(audioData)
	body.WriteString("\r\n")

	// Write model part
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
	body.WriteString("whisper-1\r\n")

	// Close boundary
	body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
	req.Header.Set("Authorization", "Bearer "+s.openAIKey)

	client := &http.Client{Timeout: 120 * time.Second} // Whisper cần timeout cao hơn
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Whisper API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Whisper API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var whisperResp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &whisperResp); err != nil {
		return "", fmt.Errorf("failed to parse Whisper response: %w", err)
	}

	config.Log.Infof("🎤 Whisper transcription successful: %d chars", len(whisperResp.Text))

	return whisperResp.Text, nil
}

// callOpenAI - Helper function gọi OpenAI Chat Completion API
func (s *aiGradingService) callOpenAI(systemPrompt, userPrompt string) (string, error) {
	reqBody := openAIRequest{
		Model: "gpt-3.5-turbo", // Changed from gpt-4 to gpt-3.5-turbo (free tier)
		Messages: []aiMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3, // Lower temperature for more consistent grading
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openAIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	config.Log.Infof("🤖 OpenAI tokens used: %d", openAIResp.Usage.TotalTokens)

	return openAIResp.Choices[0].Message.Content, nil
}

// callGemini - Helper function gọi Google Gemini API
func (s *aiGradingService) callGemini(systemPrompt, userPrompt string) (string, error) {
	// Gemini không có system prompt riêng, gộp vào user prompt
	combinedPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: combinedPrompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Gemini API endpoint - sử dụng v1beta với model gemini-2.0-flash-exp (FREE)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", s.geminiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Gemini API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	config.Log.Infof("🤖 Gemini tokens used: %d", geminiResp.UsageMetadata.TotalTokenCount)

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// cleanJSONResponse removes markdown code blocks from JSON responses
// Gemini often wraps JSON in ```json...``` markers, this function extracts the actual JSON
func cleanJSONResponse(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove ```json prefix if present
	if strings.HasPrefix(input, "```json") {
		input = strings.TrimPrefix(input, "```json")
		input = strings.TrimSpace(input)
	} else if strings.HasPrefix(input, "```") {
		input = strings.TrimPrefix(input, "```")
		input = strings.TrimSpace(input)
	}

	// Remove ``` suffix if present
	if strings.HasSuffix(input, "```") {
		input = strings.TrimSuffix(input, "```")
		input = strings.TrimSpace(input)
	}

	return input
}
