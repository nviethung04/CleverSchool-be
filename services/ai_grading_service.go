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
	GradeImageAnswer(imageURLs []string, rubric string) (*ImageGradeResult, error)
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

// ImageGradeResult - Kết quả chấm ảnh (bài làm viết tay, vẽ, diagram, etc.)
type ImageGradeResult struct {
	OverallScore      float32  `json:"overall_score"`      // Điểm tổng (0-10)
	ContentScore      float32  `json:"content_score"`      // Điểm nội dung
	AccuracyScore     float32  `json:"accuracy_score"`     // Điểm độ chính xác
	PresentationScore float32  `json:"presentation_score"` // Điểm trình bày
	Feedback          string   `json:"feedback"`           // Nhận xét chi tiết
	Strengths         []string `json:"strengths"`          // Điểm mạnh
	Improvements      []string `json:"improvements"`       // Cần cải thiện
	RecognizedText    string   `json:"recognized_text"`    // Text nhận diện được từ ảnh (nếu có)
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

// GradeImageAnswer - Chấm bài từ ảnh (bài viết tay, vẽ, diagram, etc.)
func (s *aiGradingService) GradeImageAnswer(imageURLs []string, rubric string) (*ImageGradeResult, error) {
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("no image URLs provided")
	}

	config.Log.Infof("🖼️ Starting image grading for %d image(s)", len(imageURLs))

	// Ưu tiên Gemini (free), fallback to OpenAI Vision
	if s.useGemini && s.geminiKey != "" {
		return s.gradeImageWithGemini(imageURLs, rubric)
	}

	if s.openAIKey != "" {
		return s.gradeImageWithOpenAI(imageURLs, rubric)
	}

	return nil, fmt.Errorf("no AI service available for image grading (need Gemini or OpenAI key)")
}

// gradeImageWithGemini - Chấm ảnh bằng Gemini Vision (hỗ trợ nhiều ảnh)
func (s *aiGradingService) gradeImageWithGemini(imageURLs []string, rubric string) (*ImageGradeResult, error) {
	config.Log.Info("🤖 Using Gemini Vision for image grading")

	// Download and encode all images
	var imageBase64List []string
	for i, imageURL := range imageURLs {
		config.Log.Infof("📥 Downloading image %d/%d: %s", i+1, len(imageURLs), imageURL)
		imageData, err := s.downloadImage(imageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download image %d: %w", i+1, err)
		}
		imageBase64 := base64.StdEncoding.EncodeToString(imageData)
		imageBase64List = append(imageBase64List, imageBase64)
	}

	systemPrompt := `You are an expert teacher grading student work from image(s).

CRITICAL REQUIREMENT: You MUST provide DETAILED, SPECIFIC feedback, not generic comments.

When evaluating:
1. Content accuracy (40%) - Point out SPECIFIC errors, wrong answers, or missing information
2. Presentation quality (30%) - Identify EXACT issues (messy handwriting, unclear diagrams, etc.)
3. Understanding demonstration (30%) - Show WHERE the student demonstrates understanding or lacks it

IMPORTANT RULES:
- Provide ALL feedback in Vietnamese language
- Be SPECIFIC: Instead of "có một vài lỗi", say "Câu 3 sai: viết 'mặt trời' thành 'mất trời'"
- Point out EXACT locations: "Dòng 2", "Câu hỏi số 4", "Phần giải phương trình"
- If score is not 10/10, you MUST explain WHY with specific examples
- List actual mistakes found, not just "có lỗi chính tả"
`

	imageContext := ""
	if len(imageURLs) > 1 {
		imageContext = fmt.Sprintf("(Student submitted %d images. Please analyze all of them together.)\n\n", len(imageURLs))
	}

	userPrompt := fmt.Sprintf(`%s

%sQuestion/Topic: %s

Analyze the image(s) carefully and provide DETAILED, SPECIFIC evaluation.

Return your response as a JSON object with this EXACT structure:
{
  "overall_score": <float 0-10>,
  "content_score": <float 0-10>,
  "accuracy_score": <float 0-10>,
  "presentation_score": <float 0-10>,
  "feedback": "<DETAILED feedback in Vietnamese with SPECIFIC examples of errors/issues>",
  "strengths": ["<specific strength 1 with example>", "<specific strength 2 with example>"],
  "improvements": ["<specific improvement 1: what to fix and where>", "<specific improvement 2: what to fix and where>"],
  "recognized_text": "<any text you can read from the image(s), empty string if none>"
}

Example of GOOD detailed feedback:
"Bài làm đạt 7/10. Câu 1 và 2 đúng hoàn toàn. Câu 3 sai: học sinh viết '2 + 2 = 5' thay vì '2 + 2 = 4'. Câu 4 thiếu bước giải thích công thức. Chữ viết tay ở dòng 5-7 khó đọc, cần viết rõ ràng hơn."

Example of BAD generic feedback (DO NOT DO THIS):
"Bài làm khá tốt. Có một vài lỗi nhỏ. Cần cải thiện trình bày."
`, systemPrompt, imageContext, rubric)

	// Call Gemini với multiple images
	responseText, err := s.callGeminiWithMultipleImages(userPrompt, imageBase64List)
	if err != nil {
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}

	config.Log.Infof("📝 Raw Gemini response (first 200 chars): %s", truncateString(responseText, 200))

	// Clean response (remove markdown code blocks)
	cleanedResponse := cleanJSONResponse(responseText)
	config.Log.Infof("🧹 Cleaned response (first 200 chars): %s", truncateString(cleanedResponse, 200))

	// Parse JSON response
	var result ImageGradeResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		config.Log.Errorf("❌ Failed to parse JSON response: %v", err)
		config.Log.Errorf("Response was: %s", cleanedResponse)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	config.Log.Infof("✅ Image grading completed - Score: %.1f/10", result.OverallScore)

	return &result, nil
}

// downloadImage - Download image from URL
func (s *aiGradingService) downloadImage(url string) ([]byte, error) {
	config.Log.Infof("📥 Downloading image from: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	config.Log.Infof("✅ Image downloaded: %d bytes", len(data))
	return data, nil
}

// callGeminiWithImage - Call Gemini API with image input
func (s *aiGradingService) callGeminiWithImage(prompt string, imageBase64 string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", s.geminiKey)

	// Gemini multipart request với image
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
					{
						"inline_data": map[string]string{
							"mime_type": "image/jpeg",
							"data":      imageBase64,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
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

// callGeminiWithMultipleImages - Call Gemini API với multiple images
func (s *aiGradingService) callGeminiWithMultipleImages(prompt string, imageBase64List []string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", s.geminiKey)

	// Build parts array: text + multiple images
	parts := []map[string]interface{}{
		{
			"text": prompt,
		},
	}

	// Add all images
	for _, imageBase64 := range imageBase64List {
		parts = append(parts, map[string]interface{}{
			"inline_data": map[string]string{
				"mime_type": "image/jpeg",
				"data":      imageBase64,
			},
		})
	}

	// Gemini multipart request với multiple images
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": parts,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
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

// gradeImageWithOpenAI - Chấm ảnh bằng OpenAI Vision (GPT-4 Vision)
func (s *aiGradingService) gradeImageWithOpenAI(imageURLs []string, rubric string) (*ImageGradeResult, error) {
	config.Log.Info("🤖 Using OpenAI Vision (GPT-4o) for image grading")

	systemPrompt := `You are an expert teacher grading student work from image(s).

CRITICAL REQUIREMENT: You MUST provide DETAILED, SPECIFIC feedback, not generic comments.

Analyze ALL provided images together and evaluate the student's answer based on:
1. Content accuracy (40%) - How correct and complete is the answer?
2. Presentation quality (30%) - Is it neat, organized, and easy to read?
3. Understanding demonstration (30%) - Does it show proper understanding?

Score on a 0-10 scale.

FEEDBACK REQUIREMENTS:
- Point out EXACT locations: "Dòng 2", "Câu hỏi số 4", "Bức ảnh thứ 2"
- Cite SPECIFIC errors: Instead of "có một vài lỗi", say "Câu 3 sai: viết mặt trời thành mất trời"
- Give CONCRETE examples: Instead of "nên cải thiện", say "Dòng 5: thiếu dấu chấm cuối câu"
- If score is not 10/10, you MUST explain WHY with specific examples

BAD feedback (generic): "Bài làm khá tốt. Phần lớn các câu hỏi đã được trả lời đúng. Tuy nhiên, vẫn có một vài lỗi nhỏ cần sửa."
GOOD feedback (specific): "Câu 1-3: Đúng. Câu 4: Sai - viết '2+2=5' thay vì '2+2=4'. Câu 5: Thiếu đơn vị đo (cm)."

IMPORTANT: Provide ALL feedback text in Vietnamese language.
`

	imageContext := ""
	if len(imageURLs) > 1 {
		imageContext = fmt.Sprintf("(Học sinh nộp %d ảnh. Hãy phân tích tất cả cùng nhau.)\n\n", len(imageURLs))
	}

	userPrompt := fmt.Sprintf(`%s

%sĐề bài/Chủ đề: %s

Hãy phân tích hình ảnh và đưa ra đánh giá CHI TIẾT.

NHỚ: Feedback phải CỤ THỂ với vị trí và ví dụ CHÍNH XÁC.

Trả về JSON object với cấu trúc SAU (không có markdown code blocks):
{
  "overall_score": <float 0-10>,
  "content_score": <float 0-10>,
  "accuracy_score": <float 0-10>,
  "presentation_score": <float 0-10>,
  "feedback": "<phản hồi chi tiết bằng tiếng Việt - PHẢI CỤ THỂ>",
  "strengths": ["<điểm mạnh cụ thể 1>", "<điểm mạnh cụ thể 2>"],
  "improvements": ["<cần cải thiện cụ thể 1 với vị trí chính xác>", "<cần cải thiện cụ thể 2 với vị trí chính xác>"],
  "recognized_text": "<văn bản đọc được từ ảnh, để trống nếu không có>"
}`, systemPrompt, imageContext, rubric)

	// Build content array with text + multiple images
	contentArray := []map[string]interface{}{
		{
			"type": "text",
			"text": userPrompt,
		},
	}

	// Add all images
	for _, imageURL := range imageURLs {
		contentArray = append(contentArray, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": imageURL,
			},
		})
	}

	// OpenAI Vision API request structure
	requestBody := map[string]interface{}{
		"model": "gpt-4o", // GPT-4 Omni có vision capability tốt hơn gpt-4-vision-preview
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": contentArray,
			},
		},
		"max_tokens": 1500,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.openAIKey))

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		config.Log.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))

		// Parse error response để lấy chi tiết
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			if resp.StatusCode == 429 {
				return nil, fmt.Errorf("OpenAI rate limit exceeded: %s. Please try again later or check your quota at https://platform.openai.com/usage", errorResp.Error.Message)
			}
			return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, errorResp.Error.Message)
		}

		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	responseText := openAIResp.Choices[0].Message.Content
	config.Log.Infof("📝 Raw OpenAI response (first 200 chars): %s", truncateString(responseText, 200))
	config.Log.Infof("🎯 OpenAI tokens used: %d", openAIResp.Usage.TotalTokens)

	// Clean response (OpenAI cũng có thể trả về markdown code blocks)
	cleanedResponse := cleanJSONResponse(responseText)
	config.Log.Infof("🧹 Cleaned response (first 200 chars): %s", truncateString(cleanedResponse, 200))

	// Parse JSON response
	var result ImageGradeResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		config.Log.Errorf("❌ Failed to parse JSON response: %v", err)
		config.Log.Errorf("Response was: %s", cleanedResponse)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	config.Log.Infof("✅ Image grading completed - Score: %.1f/10", result.OverallScore)

	return &result, nil
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
