package main

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// SeedFlashcards creates fake flashcard data
func (s *Seeder) SeedFlashcards() {
	fmt.Println("🎯 Creating Flashcard fake data...")

	// Sample vocabulary data
	vocabularies := []models.Vocabulary{
		{
			Word:               "hello",
			Phonetic:           "/həˈloʊ/",
			Translation:        "xin chào",
			Definition:         "A greeting used when meeting someone",
			PartOfSpeech:       "interjection",
			Level:              "beginner",
			Tags:               "greeting,basic,daily",
			Difficulty:         1,
			Frequency:          100,
			ExampleSentence:    "Hello, how are you today?",
			ExampleTranslation: "Xin chào, hôm nay bạn khỏe không?",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "goodbye",
			Phonetic:           "/ɡʊdˈbaɪ/",
			Translation:        "tạm biệt",
			Definition:         "A farewell expression used when parting",
			PartOfSpeech:       "interjection",
			Level:              "beginner",
			Tags:               "greeting,basic,daily",
			Difficulty:         1,
			Frequency:          90,
			ExampleSentence:    "Goodbye, see you tomorrow!",
			ExampleTranslation: "Tạm biệt, hẹn gặp lại ngày mai!",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "thank you",
			Phonetic:           "/θæŋk juː/",
			Translation:        "cám ơn",
			Definition:         "An expression of gratitude",
			PartOfSpeech:       "phrase",
			Level:              "beginner",
			Tags:               "politeness,basic,daily",
			Difficulty:         1,
			Frequency:          95,
			ExampleSentence:    "Thank you for your help!",
			ExampleTranslation: "Cám ơn bạn đã giúp đỡ!",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "please",
			Phonetic:           "/pliːz/",
			Translation:        "xin vui lòng",
			Definition:         "Used to make a request more polite",
			PartOfSpeech:       "adverb",
			Level:              "beginner",
			Tags:               "politeness,basic,daily",
			Difficulty:         1,
			Frequency:          85,
			ExampleSentence:    "Please help me with this task.",
			ExampleTranslation: "Xin vui lòng giúp tôi với công việc này.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "beautiful",
			Phonetic:           "/ˈbjuːtɪfʊl/",
			Translation:        "đẹp",
			Definition:         "Pleasing to look at; attractive",
			PartOfSpeech:       "adjective",
			Level:              "intermediate",
			Tags:               "description,appearance,emotion",
			Difficulty:         2,
			Frequency:          70,
			ExampleSentence:    "The sunset is beautiful tonight.",
			ExampleTranslation: "Hoàng hôn tối nay thật đẹp.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "wonderful",
			Phonetic:           "/ˈwʌndəfʊl/",
			Translation:        "tuyệt vời",
			Definition:         "Extremely good; excellent",
			PartOfSpeech:       "adjective",
			Level:              "intermediate",
			Tags:               "emotion,positive,description",
			Difficulty:         2,
			Frequency:          65,
			ExampleSentence:    "We had a wonderful time at the party.",
			ExampleTranslation: "Chúng tôi đã có khoảng thời gian tuyệt vời tại bữa tiệc.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "extraordinary",
			Phonetic:           "/ɪkˈstrɔːdɪnəri/",
			Translation:        "phi thường",
			Definition:         "Very unusual or remarkable",
			PartOfSpeech:       "adjective",
			Level:              "advanced",
			Tags:               "description,academic,formal",
			Difficulty:         4,
			Frequency:          30,
			ExampleSentence:    "She has extraordinary talent in music.",
			ExampleTranslation: "Cô ấy có tài năng phi thường về âm nhạc.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "sophisticated",
			Phonetic:           "/səˈfɪstɪkeɪtɪd/",
			Translation:        "tinh tế, phức tạp",
			Definition:         "Having a refined knowledge of culture and fashion",
			PartOfSpeech:       "adjective",
			Level:              "advanced",
			Tags:               "description,culture,academic",
			Difficulty:         4,
			Frequency:          25,
			ExampleSentence:    "The restaurant has a sophisticated atmosphere.",
			ExampleTranslation: "Nhà hàng có không khí tinh tế.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "happy",
			Phonetic:           "/ˈhæpi/",
			Translation:        "vui vẻ, hạnh phúc",
			Definition:         "Feeling or showing pleasure or contentment",
			PartOfSpeech:       "adjective",
			Level:              "beginner",
			Tags:               "emotion,basic,daily",
			Difficulty:         1,
			Frequency:          80,
			ExampleSentence:    "I am happy to see you again.",
			ExampleTranslation: "Tôi rất vui khi gặp lại bạn.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "intelligent",
			Phonetic:           "/ɪnˈtelɪdʒənt/",
			Translation:        "thông minh",
			Definition:         "Having the ability to learn and understand quickly",
			PartOfSpeech:       "adjective",
			Level:              "intermediate",
			Tags:               "personality,description,academic",
			Difficulty:         3,
			Frequency:          55,
			ExampleSentence:    "She is very intelligent and hardworking.",
			ExampleTranslation: "Cô ấy rất thông minh và chăm chỉ.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
	}

	// Create vocabularies
	for i := range vocabularies {
		if err := db.MasterDB.Create(&vocabularies[i]).Error; err != nil {
			log.Printf("Failed to create vocabulary %s: %v", vocabularies[i].Word, err)
			continue
		}
		fmt.Printf("✅ Created vocabulary: %s\n", vocabularies[i].Word)
	}

	// Get some lessons to link vocabularies
	var lessons []models.Lesson
	if err := db.MasterDB.Limit(3).Find(&lessons).Error; err != nil {
		log.Printf("Failed to get lessons: %v", err)
		return
	}

	if len(lessons) == 0 {
		fmt.Println("⚠️ No lessons found in database. Please create some lessons first.")
		return
	}

	// Create lesson-vocabulary relationships
	fmt.Println("🔗 Linking vocabularies to lessons...")

	for i, lesson := range lessons {
		// Each lesson gets 3-4 vocabularies
		startIdx := i * 3
		endIdx := startIdx + 3
		if endIdx > len(vocabularies) {
			endIdx = len(vocabularies)
		}

		for j, vocab := range vocabularies[startIdx:endIdx] {
			lessonVocab := models.LessonVocabulary{
				LessonID:     lesson.ID,
				VocabularyID: vocab.ID,
				SortOrder:    j + 1,
				IsRequired:   true,
			}

			if err := db.MasterDB.Create(&lessonVocab).Error; err != nil {
				log.Printf("Failed to link vocab %s to lesson %s: %v", vocab.Word, lesson.Title, err)
				continue
			}
			fmt.Printf("✅ Linked %s to lesson: %s\n", vocab.Word, lesson.Title)
		}
	}

	// Get some students to create progress
	var students []models.User
	if err := db.MasterDB.
		Joins("LEFT JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("LEFT JOIN roles ON roles.id = urr.role_id").
		Where("roles.name ILIKE ?", "%student%").
		Limit(5).
		Find(&students).Error; err != nil {

		log.Printf("Failed to get students: %v", err)

		if err := db.MasterDB.Limit(5).Find(&students).Error; err != nil {
			log.Printf("Failed to get any users: %v", err)
			return
		}
	}

	if len(students) == 0 {
		fmt.Println("⚠️ No students found in database. Skipping progress creation.")
		return
	}

	// Create student progress data
	fmt.Println("📈 Creating student progress data...")

	statuses := []string{"new", "learning", "known", "mastered"}

	for _, student := range students {
		for _, vocab := range vocabularies {
			// Skip some vocabularies for variety
			if rand.Intn(3) == 0 {
				continue
			}

			status := statuses[rand.Intn(len(statuses))]
			isKnown := status == "known" || status == "mastered"

			progress := models.StudentVocabularyProgress{
				StudentID:    student.ID,
				VocabularyID: vocab.ID,
				LessonID:     lessons[0].ID, // Use first lesson
				Status:       status,
				IsKnown:      isKnown,
				StudyCount:   rand.Intn(10) + 1,
				CorrectCount: rand.Intn(5) + 1,
			}

			// Add some timestamps
			if status != "new" {
				lastStudied := time.Now().AddDate(0, 0, -rand.Intn(7))
				progress.LastStudiedAt = &lastStudied
			}

			if status == "mastered" {
				mastered := time.Now().AddDate(0, 0, -rand.Intn(14))
				progress.MasteredAt = &mastered
			}

			// Add pronunciation scores for some records
			if rand.Intn(2) == 0 {
				score := 60.0 + rand.Float64()*35     // 60-95 range
				bestScore := score + rand.Float64()*5 // slightly higher
				progress.PronunciationScore = &score
				progress.BestPronunciationScore = &bestScore
				progress.PronunciationAttempts = rand.Intn(8) + 1
			}

			if err := db.MasterDB.Create(&progress).Error; err != nil {
				log.Printf("Failed to create progress for student %d, vocab %s: %v",
					student.ID, vocab.Word, err)
				continue
			}
		}
		fmt.Printf("✅ Created progress for student: %s\n", student.Name)
	}

	// Create some flashcard sessions
	fmt.Println("🎮 Creating flashcard sessions...")

	for _, student := range students[:3] { // Only first 3 students
		for _, lesson := range lessons {
			// Skip some combinations
			if rand.Intn(3) == 0 {
				continue
			}

			session := models.FlashcardSession{
				StudentID:      student.ID,
				LessonID:       lesson.ID,
				TotalCards:     rand.Intn(10) + 5,        // 5-15 cards
				CompletedCards: rand.Intn(8) + 2,         // 2-10 completed
				KnownCards:     rand.Intn(5) + 1,         // 1-6 known
				StudyDuration:  rand.Intn(1200) + 300,    // 5-25 minutes in seconds
				SessionScore:   60.0 + rand.Float64()*35, // 60-95 score
			}

			// Add timestamps
			startTime := time.Now().AddDate(0, 0, -rand.Intn(30)) // Within last 30 days
			session.StartedAt = &startTime

			if rand.Intn(2) == 0 { // 50% chance session is completed
				completedTime := startTime.Add(time.Duration(session.StudyDuration) * time.Second)
				session.CompletedAt = &completedTime
			}

			if err := db.MasterDB.Create(&session).Error; err != nil {
				log.Printf("Failed to create session for student %d, lesson %d: %v",
					student.ID, lesson.ID, err)
				continue
			}

			// Create some activities for this session
			activities := []string{"view", "flip", "mark_known", "mark_unknown", "pronunciation"}
			for k := 0; k < rand.Intn(5)+2; k++ { // 2-7 activities
				if len(vocabularies) <= k {
					break
				}

				activity := models.FlashcardActivity{
					SessionID:    session.ID,
					VocabularyID: vocabularies[k].ID,
					ActivityType: activities[rand.Intn(len(activities))],
					ResponseTime: rand.Intn(5000) + 500, // 0.5-5.5 seconds
				}

				if activity.ActivityType == "mark_known" || activity.ActivityType == "mark_unknown" {
					if activity.ActivityType == "mark_known" {
						activity.Response = "known"
					} else {
						activity.Response = "unknown"
					}
				}

				if activity.ActivityType == "pronunciation" {
					score := 60.0 + rand.Float64()*35
					activity.PronunciationScore = &score
					activity.Response = "pronunciation_recorded"
				}

				if err := db.MasterDB.Create(&activity).Error; err != nil {
					log.Printf("Failed to create activity: %v", err)
					continue
				}
			}

			fmt.Printf("✅ Created session for student %s, lesson %s\n",
				student.Name, lesson.Title)
		}
	}

	fmt.Println("🎉 Flashcard fake data creation completed!")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   - %d vocabularies created\n", len(vocabularies))
	fmt.Printf("   - Linked to %d lessons\n", len(lessons))
	fmt.Printf("   - Progress for %d students\n", len(students))
	fmt.Printf("   - Multiple study sessions and activities\n")
}
