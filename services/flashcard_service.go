package services

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type FlashcardService struct {
	db *gorm.DB
}

func NewFlashcardService() *FlashcardService {
	return &FlashcardService{
		db: db.MasterDB,
	}
}

// ========== Vocabulary Management ==========

// CreateVocabulary creates a new vocabulary
func (s *FlashcardService) CreateVocabulary(req prot.CreateVocabularyRequest, createdBy int64) (*dto.VocabularyDTO, error) {
	wordAudioId := int64(req.WordAudioId)
	vocabulary := models.Vocabulary{
		Word:               req.Word,
		Phonetic:           req.Phonetic,
		Translation:        req.Translation,
		Definition:         req.Definition,
		PartOfSpeech:       req.PartOfSpeech,
		Level:              req.Level,
		Tags:               req.Tags,
		Difficulty:         int(req.Difficulty),
		Frequency:          int(req.Frequency),
		WordAudioID:        &wordAudioId,
		ImageID:            &req.ImageId,
		ExampleSentence:    req.ExampleSentence,
		ExampleTranslation: req.ExampleTranslation,
		ExampleAudioID:     &req.ExampleAudioId,
		CreatedBy:          createdBy,
		UpdatedBy:          createdBy,
	}

	if err := s.db.Create(&vocabulary).Error; err != nil {
		return nil, fmt.Errorf("failed to create vocabulary: %w", err)
	}

	return s.GetVocabularyByID(vocabulary.ID)
}

// GetVocabularyByID retrieves vocabulary by ID
func (s *FlashcardService) GetVocabularyByID(id int64) (*dto.VocabularyDTO, error) {
	var vocabulary models.Vocabulary

	err := s.db.Preload("WordAudio").
		Preload("Image").
		Preload("ExampleAudio").
		First(&vocabulary, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vocabulary not found")
		}
		return nil, fmt.Errorf("failed to get vocabulary: %w", err)
	}

	return s.mapVocabularyToDTO(&vocabulary), nil
}

// UpdateVocabulary updates an existing vocabulary
func (s *FlashcardService) UpdateVocabulary(id int64, req prot.UpdateVocabularyRequest, updatedBy int64) (*dto.VocabularyDTO, error) {
	var vocabulary models.Vocabulary

	if err := s.db.First(&vocabulary, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vocabulary not found")
		}
		return nil, fmt.Errorf("failed to find vocabulary: %w", err)
	}

	// Update fields if provided
	if req.Word != nil {
		vocabulary.Word = req.Word.Value
	}
	if req.Translation != nil {
		vocabulary.Translation = req.Translation.Value
	}
	if req.Level != nil {
		vocabulary.Level = req.Level.Value
	}
	// ... update other fields

	vocabulary.UpdatedBy = updatedBy

	if err := s.db.Save(&vocabulary).Error; err != nil {
		return nil, fmt.Errorf("failed to update vocabulary: %w", err)
	}

	return s.GetVocabularyByID(id)
}

// ========== Lesson Vocabulary Management ==========

// AddVocabulariesToLesson adds vocabularies to a lesson
func (s *FlashcardService) AddVocabulariesToLesson(lessonID int64, req prot.AddVocabularyToLessonRequest) error {
	for i, vocabID := range req.VocabularyIds {
		lessonVocab := models.LessonVocabulary{
			LessonID:     lessonID,
			VocabularyID: vocabID,
			SortOrder:    int(req.SortOrder) + i,
			IsRequired:   req.IsRequired,
		}

		if err := s.db.Create(&lessonVocab).Error; err != nil {
			return fmt.Errorf("failed to add vocabulary %d to lesson: %w", vocabID, err)
		}
	}

	return nil
}

// SyncLessonVocabularies replaces all vocabulary links for a lesson.
func (s *FlashcardService) SyncLessonVocabularies(lessonID int64, vocabularyIDs []int64) error {
	tx := s.db.Begin()

	if err := tx.Unscoped().Where("lesson_id = ?", lessonID).Delete(&models.LessonVocabulary{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear lesson vocabularies: %w", err)
	}

	for i, vocabID := range vocabularyIDs {
		if vocabID <= 0 {
			continue
		}
		lessonVocab := models.LessonVocabulary{
			LessonID:     lessonID,
			VocabularyID: vocabID,
			SortOrder:    i + 1,
			IsRequired:   true,
		}
		if err := tx.Create(&lessonVocab).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to sync vocabulary %d to lesson: %w", vocabID, err)
		}
	}

	return tx.Commit().Error
}

// GetLessonVocabularies retrieves all vocabularies for a lesson
func (s *FlashcardService) GetLessonVocabularies(lessonID int64, studentID *int64) (*dto.FlashcardLessonDTO, error) {
	var lesson models.Lesson
	if err := s.db.First(&lesson, lessonID).Error; err != nil {
		return nil, fmt.Errorf("lesson not found: %w", err)
	}

	var lessonVocabs []models.LessonVocabulary
	query := s.db.Preload("Vocabulary").
		Preload("Vocabulary.WordAudio").
		Preload("Vocabulary.Image").
		Preload("Vocabulary.ExampleAudio").
		Where("lesson_id = ?", lessonID).
		Order("sort_order ASC")

	if err := query.Find(&lessonVocabs).Error; err != nil {
		return nil, fmt.Errorf("failed to get lesson vocabularies: %w", err)
	}

	// Get progress if student ID provided
	var progressMap map[int64]*models.StudentVocabularyProgress
	if studentID != nil {
		progressMap = s.getStudentProgressMap(*studentID, lessonID)
	}

	// Convert to DTOs
	vocabularyDTOs := make([]dto.LessonVocabularyDTO, len(lessonVocabs))
	for i, lv := range lessonVocabs {
		vocabularyDTOs[i] = dto.LessonVocabularyDTO{
			ID:           lv.ID,
			LessonID:     lv.LessonID,
			VocabularyID: lv.VocabularyID,
			SortOrder:    lv.SortOrder,
			IsRequired:   lv.IsRequired,
			Vocabulary:   s.mapVocabularyToDTO(&lv.Vocabulary),
		}

		// Add progress if available
		if progress, exists := progressMap[lv.VocabularyID]; exists {
			vocabularyDTOs[i].Progress = s.mapProgressToDTO(progress)
		}
	}

	return &dto.FlashcardLessonDTO{
		ID:           lesson.ID,
		Title:        lesson.Title,
		Description:  lesson.Description,
		Vocabularies: vocabularyDTOs,
	}, nil
}

// ========== Flashcard Session Management ==========

// StartFlashcardSession starts a new flashcard session
func (s *FlashcardService) StartFlashcardSession(studentID int64, req prot.StartFlashcardSessionRequest) (map[string]interface{}, error) {
	// Step 1: Check if there's an active session for this lesson
	var existingSession models.FlashcardSession
	existingSessionFound := s.db.Where("student_id = ? AND lesson_id = ? AND completed_at IS NULL", studentID, req.LessonId).
		Order("created_at DESC").First(&existingSession).Error == nil

	if existingSessionFound {
		// Resume existing session
		sessionData, err := s.ResumeFlashcardSession(studentID, existingSession.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to resume existing session: %w", err)
		}

		// Add session status info
		sessionData["is_new_session"] = false
		sessionData["message"] = "Resumed existing session"

		return sessionData, nil
	}

	// Step 2: Create new session
	session := models.FlashcardSession{
		StudentID: studentID,
		LessonID:  req.LessonId,
		StartedAt: &time.Time{},
	}
	*session.StartedAt = time.Now()

	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	// Step 3: Get vocabularies for new session
	var vocabularies []models.Vocabulary
	query := s.db.Table("vocabularies v").
		Select("v.*").
		Joins("JOIN lesson_vocabularies lv ON lv.vocabulary_id = v.id").
		Where("lv.lesson_id = ?", req.LessonId).
		Preload("WordAudio").
		Preload("Image").
		Preload("ExampleAudio").
		Order("v.difficulty ASC, v.id ASC")

	if err := query.Find(&vocabularies).Error; err != nil {
		return nil, fmt.Errorf("failed to get vocabularies: %w", err)
	}

	// Ensure we have vocabularies
	if len(vocabularies) == 0 {
		return nil, fmt.Errorf("no vocabularies found for this lesson")
	}

	// Update session with correct total cards count
	session.TotalCards = len(vocabularies)
	s.db.Save(&session)

	// Get progress for vocabularies
	progressMap := s.getStudentProgressMap(studentID, req.LessonId)

	var vocabularyDTOs []map[string]interface{}
	for _, vocab := range vocabularies {
		vocabDTO := s.mapVocabularyToDTO(&vocab)

		vocabData := map[string]interface{}{
			"vocabulary": vocabDTO,
		}

		if progress, exists := progressMap[vocab.ID]; exists {
			vocabData["progress"] = s.mapProgressToDTO(progress)
		}

		vocabularyDTOs = append(vocabularyDTOs, vocabData)
	}

	result := map[string]interface{}{
		"session_id":      session.ID,
		"session":         s.mapSessionToDTO(&session),
		"vocabularies":    vocabularyDTOs,
		"current_index":   0, // New session starts from beginning
		"total_cards":     len(vocabularyDTOs),
		"completed_cards": 0,
		"remaining_cards": len(vocabularyDTOs),
		"lesson_id":       req.LessonId,
		"is_new_session":  true,
		"message":         "Started new session",
	}

	return result, nil
}

// GetAllVocabularies gets all vocabularies with pagination and filtering
func (s *FlashcardService) GetAllVocabularies(filters map[string]interface{}) ([]dto.VocabularyDTO, int64, error) {

	page := filters["page"].(int)
	limit := filters["limit"].(int)
	search := filters["search"].(string)
	difficulty := filters["difficulty"].(int)
	partOfSpeech := filters["part_of_speech"].(string)

	offset := (page - 1) * limit

	query := s.db.Model(&models.Vocabulary{})

	if search != "" {
		query = query.Where("word ILIKE ? OR translation ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if difficulty > 0 {
		query = query.Where("difficulty = ?", difficulty)
	}
	if partOfSpeech != "" {
		query = query.Where("part_of_speech = ?", partOfSpeech)
	}

	var total int64
	query.Count(&total)

	var vocabularies []models.Vocabulary
	err := query.Preload("WordAudio").Preload("Image").Preload("ExampleAudio").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&vocabularies).Error
	if err != nil {
		return nil, 0, err
	}

	var vocabularyDTOs []dto.VocabularyDTO
	for _, vocab := range vocabularies {
		dtoObj := s.mapVocabularyToDTO(&vocab)
		if dtoObj != nil {
			vocabularyDTOs = append(vocabularyDTOs, *dtoObj)
		}
	}

	return vocabularyDTOs, total, nil
}

// RecordFlashcardActivity records a flashcard activity
func (s *FlashcardService) RecordFlashcardActivity(sessionID int64, vocabularyID int64, req prot.RecordFlashcardActivityRequest) error {
	pronunciationScore := float64(req.PronunciationScore.Value)
	activity := models.FlashcardActivity{
		SessionID:          sessionID,
		VocabularyID:       vocabularyID,
		ActivityType:       req.ActivityType,
		Response:           req.Response,
		ResponseTime:       int(req.ResponseTime),
		PronunciationScore: &pronunciationScore,
	}

	return s.db.Create(&activity).Error
}

// UpdateVocabularyProgress updates student's progress with a vocabulary
func (s *FlashcardService) UpdateVocabularyProgress(studentID, vocabularyID int64, req prot.UpdateVocabularyProgressRequest) error {
	// First, check if vocabulary exists
	var vocabulary models.Vocabulary
	if err := s.db.First(&vocabulary, vocabularyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("vocabulary not found")
		}
		return fmt.Errorf("failed to find vocabulary: %w", err)
	}

	// Get lesson ID from lesson_vocabularies table
	var lessonVocab models.LessonVocabulary
	if err := s.db.Where("vocabulary_id = ?", vocabularyID).First(&lessonVocab).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("vocabulary not assigned to any lesson")
		}
		return fmt.Errorf("failed to find lesson vocabulary: %w", err)
	}

	var progress models.StudentVocabularyProgress

	// Find existing progress record
	result := s.db.Where("student_id = ? AND vocabulary_id = ?", studentID, vocabularyID).
		First(&progress)

	now := time.Now()

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new progress record
		progress = models.StudentVocabularyProgress{
			StudentID:     studentID,
			VocabularyID:  vocabularyID,
			LessonID:      lessonVocab.LessonID,
			Status:        req.Status,
			IsKnown:       req.IsKnown,
			StudyCount:    1,
			CorrectCount:  0,
			LastStudiedAt: &now,
		}

		if req.Status == "mastered" {
			progress.MasteredAt = &now
		}

		return s.db.Create(&progress).Error
	}

	if result.Error != nil {
		return fmt.Errorf("failed to get progress: %w", result.Error)
	}

	// Update existing progress
	progress.Status = req.Status
	progress.IsKnown = req.IsKnown
	progress.StudyCount++
	progress.LastStudiedAt = &now

	if req.Status == "mastered" && progress.MasteredAt == nil {
		progress.MasteredAt = &now
	}

	return s.db.Save(&progress).Error
}

// ========== Helper Methods ==========

func (s *FlashcardService) mapVocabularyToDTO(vocab *models.Vocabulary) *dto.VocabularyDTO {
	vocabularyDTO := &dto.VocabularyDTO{
		ID:                 vocab.ID,
		Word:               vocab.Word,
		Phonetic:           vocab.Phonetic,
		Translation:        vocab.Translation,
		Definition:         vocab.Definition,
		PartOfSpeech:       vocab.PartOfSpeech,
		Level:              vocab.Level,
		Tags:               vocab.Tags,
		Difficulty:         vocab.Difficulty,
		Frequency:          vocab.Frequency,
		ExampleSentence:    vocab.ExampleSentence,
		ExampleTranslation: vocab.ExampleTranslation,
		CreatedAt:          vocab.CreatedAt,
		UpdatedAt:          vocab.UpdatedAt,
	}

	// Map media if loaded
	if vocab.WordAudio != nil {
		vocabularyDTO.WordAudio = s.mapMediaToDTO(vocab.WordAudio)
	}
	if vocab.Image != nil {
		vocabularyDTO.Image = s.mapMediaToDTO(vocab.Image)
	}
	if vocab.ExampleAudio != nil {
		vocabularyDTO.ExampleAudio = s.mapMediaToDTO(vocab.ExampleAudio)
	}

	return vocabularyDTO
}

func (s *FlashcardService) mapMediaToDTO(media *models.Media) *dto.MediaDTO {
	return &dto.MediaDTO{
		ID:       media.ID,
		FileName: media.FileName,
		FilePath: media.FilePath,
		FullPath: media.FullPath,
		FileType: *media.FileType,
		FileSize: *media.FileSize,
		URL:      media.FullPath, // Adjust based on your URL generation logic
	}
}

func (s *FlashcardService) mapProgressToDTO(progress *models.StudentVocabularyProgress) *dto.StudentVocabularyProgressDTO {
	return &dto.StudentVocabularyProgressDTO{
		ID:                     progress.ID,
		StudentID:              progress.StudentID,
		VocabularyID:           progress.VocabularyID,
		LessonID:               progress.LessonID,
		Status:                 progress.Status,
		IsKnown:                progress.IsKnown,
		StudyCount:             progress.StudyCount,
		CorrectCount:           progress.CorrectCount,
		LastStudiedAt:          progress.LastStudiedAt,
		MasteredAt:             progress.MasteredAt,
		PronunciationScore:     progress.PronunciationScore,
		BestPronunciationScore: progress.BestPronunciationScore,
		PronunciationAttempts:  progress.PronunciationAttempts,
	}
}

func (s *FlashcardService) mapSessionToDTO(session *models.FlashcardSession) *dto.FlashcardSessionDTO {
	dto := &dto.FlashcardSessionDTO{
		ID:             session.ID,
		StudentID:      session.StudentID,
		LessonID:       session.LessonID,
		TotalCards:     session.TotalCards,
		CompletedCards: session.CompletedCards,
		KnownCards:     session.KnownCards,
		StudyDuration:  session.StudyDuration,
		SessionScore:   session.SessionScore,
		StartedAt:      session.StartedAt,
		CompletedAt:    session.CompletedAt,
		CreatedAt:      session.CreatedAt,
	}

	// Calculate progress percentage
	if session.TotalCards > 0 {
		dto.ProgressPercentage = float64(session.CompletedCards) / float64(session.TotalCards) * 100
	}

	return dto
}

func (s *FlashcardService) getStudentProgressMap(studentID, lessonID int64) map[int64]*models.StudentVocabularyProgress {
	var progresses []models.StudentVocabularyProgress

	s.db.Where("student_id = ? AND lesson_id = ?", studentID, lessonID).
		Find(&progresses)

	progressMap := make(map[int64]*models.StudentVocabularyProgress)
	for i := range progresses {
		progressMap[progresses[i].VocabularyID] = &progresses[i]
	}

	return progressMap
}

// ========== Session Management ==========

// CompleteFlashcardSession completes a flashcard session and updates lesson progress
func (s *FlashcardService) CompleteFlashcardSession(sessionID, userID int64, req prot.CompleteFlashcardSessionRequest) (*dto.FlashcardSessionDTO, error) {
	var session models.FlashcardSession
	if err := s.db.Where("id = ? AND student_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	// Update session with completion data
	now := time.Now()
	session.StudyDuration = int(req.TotalTimeSpent)
	if req.StudyDuration > 0 {
		session.StudyDuration = int(req.StudyDuration) // for backward compatibility
	}
	session.SessionScore = req.AverageScore
	if req.SessionScore > 0 {
		session.SessionScore = req.SessionScore // for backward compatibility
	}
	session.CompletedAt = &now
	session.UpdatedAt = now

	if err := s.db.Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to complete session: %w", err)
	}

	// Update lesson progress if needed
	// TODO: Add lesson progress update logic here

	return s.mapSessionToDTO(&session), nil
}

// GetLessonProgress gets current progress and last session info for a lesson
func (s *FlashcardService) GetLessonProgress(userID, lessonID int64) (map[string]interface{}, error) {
	// Get lesson vocabulary count
	var totalVocabularies int64
	if err := s.db.Model(&models.LessonVocabulary{}).Where("lesson_id = ?", lessonID).Count(&totalVocabularies).Error; err != nil {
		return nil, fmt.Errorf("failed to get total vocabularies: %w", err)
	}

	// Get studied vocabulary count
	var studiedVocabularies int64
	if err := s.db.Model(&models.StudentVocabularyProgress{}).
		Where("student_id = ? AND lesson_id = ? AND study_count > 0", userID, lessonID).
		Count(&studiedVocabularies).Error; err != nil {
		return nil, fmt.Errorf("failed to get studied vocabularies: %w", err)
	}

	// Get last active session
	var activeSession models.FlashcardSession
	activeSessionExists := false
	if err := s.db.Where("student_id = ? AND lesson_id = ? AND completed_at IS NULL", userID, lessonID).
		Order("created_at DESC").First(&activeSession).Error; err == nil {
		activeSessionExists = true
	}

	// Get last studied time
	var lastStudiedAt *time.Time
	s.db.Model(&models.StudentVocabularyProgress{}).
		Select("MAX(last_studied_at)").
		Where("student_id = ? AND lesson_id = ?", userID, lessonID).
		Scan(&lastStudiedAt)

	result := map[string]interface{}{
		"lesson_id":            lessonID,
		"total_vocabularies":   totalVocabularies,
		"studied_vocabularies": studiedVocabularies,
		"completion_rate":      float64(studiedVocabularies) / float64(totalVocabularies),
		"last_studied_at":      lastStudiedAt,
		"active_session":       nil,
	}

	if activeSessionExists {
		result["active_session"] = map[string]interface{}{
			"session_id":               activeSession.ID,
			"current_vocabulary_index": activeSession.CompletedCards,
			"status":                   "active",
		}
	}

	return result, nil
}

// GetActiveSession gets current active session for a lesson if exists
func (s *FlashcardService) GetActiveSession(userID, lessonID int64) (*dto.FlashcardSessionDTO, error) {
	var session models.FlashcardSession
	if err := s.db.Where("student_id = ? AND lesson_id = ? AND completed_at IS NULL", userID, lessonID).
		Order("created_at DESC").First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active session found")
		}
		return nil, fmt.Errorf("failed to find active session: %w", err)
	}

	return s.mapSessionToDTO(&session), nil
}

// ResumeFlashcardSession resumes an existing flashcard session
func (s *FlashcardService) ResumeFlashcardSession(userID, sessionID int64) (map[string]interface{}, error) {
	var session models.FlashcardSession
	if err := s.db.Where("id = ? AND student_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	if session.CompletedAt != nil {
		return nil, fmt.Errorf("session already completed")
	}

	// Get remaining vocabularies for this session
	var vocabularies []models.Vocabulary
	query := s.db.Table("vocabularies v").
		Select("v.*").
		Joins("JOIN lesson_vocabularies lv ON lv.vocabulary_id = v.id").
		Where("lv.lesson_id = ?", session.LessonID).
		Preload("WordAudio").
		Preload("Image").
		Preload("ExampleAudio").
		Order("v.difficulty ASC, v.id ASC")

	if err := query.Find(&vocabularies).Error; err != nil {
		return nil, fmt.Errorf("failed to get vocabularies: %w", err)
	}

	// Get progress for vocabularies
	progressMap := s.getStudentProgressMap(userID, session.LessonID)

	var vocabularyDTOs []map[string]interface{}
	for _, vocab := range vocabularies {
		vocabDTO := s.mapVocabularyToDTO(&vocab)

		vocabData := map[string]interface{}{
			"vocabulary": vocabDTO,
		}

		if progress, exists := progressMap[vocab.ID]; exists {
			vocabData["progress"] = s.mapProgressToDTO(progress)
		}

		vocabularyDTOs = append(vocabularyDTOs, vocabData)
	}

	// Ensure we have vocabularies
	if len(vocabularyDTOs) == 0 {
		return nil, fmt.Errorf("no vocabularies found for this lesson")
	}

	// Calculate current_index based on studied vocabularies from flashcard_activities
	var studiedVocabularies []int64
	s.db.Model(&models.FlashcardActivity{}).
		Where("session_id = ? AND activity_type = ?", session.ID, "view").
		Distinct("vocabulary_id").
		Pluck("vocabulary_id", &studiedVocabularies)

	// current_index = vị trí từ tiếp theo cần học
	// Ví dụ: đã view "hello" → current_index = 1 (để hiển thị "study")
	completedCards := len(studiedVocabularies)
	currentIndex := completedCards // Từ tiếp theo chưa học

	// Đảm bảo không vượt quá số từ có sẵn
	if currentIndex >= len(vocabularyDTOs) {
		currentIndex = len(vocabularyDTOs) - 1
	}

	remainingCards := len(vocabularyDTOs) - completedCards

	result := map[string]interface{}{
		"session_id":      session.ID,
		"session":         s.mapSessionToDTO(&session),
		"vocabularies":    vocabularyDTOs,
		"current_index":   currentIndex,
		"total_cards":     len(vocabularyDTOs),
		"completed_cards": completedCards,
		"remaining_cards": remainingCards,
		"lesson_id":       session.LessonID,
	}

	return result, nil
}
