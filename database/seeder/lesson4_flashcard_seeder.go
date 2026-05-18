package main

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"fmt"
	"log"
)

func SeedLesson4Flashcards() {
	fmt.Println("🎯 Creating Flashcard data for Lesson 4...")

	// Check if lesson 4 exists
	var lesson models.Lesson
	if err := db.MasterDB.First(&lesson, 4).Error; err != nil {
		fmt.Printf("❌ Lesson 4 not found: %v\n", err)
		return
	}
	fmt.Printf("✅ Found lesson 4: %s\n", lesson.Title)

	// Create vocabularies for lesson 4
	vocabularies := []models.Vocabulary{
		{
			Word:               "study",
			Phonetic:           "/ˈstʌdi/",
			Translation:        "học tập",
			Definition:         "To learn about something",
			PartOfSpeech:       "verb",
			Level:              "beginner",
			Tags:               "education,learning,basic",
			Difficulty:         1,
			Frequency:          90,
			ExampleSentence:    "I study English every day.",
			ExampleTranslation: "Tôi học tiếng Anh mỗi ngày.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "learn",
			Phonetic:           "/lɜːrn/",
			Translation:        "học hỏi",
			Definition:         "To gain knowledge or skill",
			PartOfSpeech:       "verb",
			Level:              "beginner",
			Tags:               "education,knowledge,basic",
			Difficulty:         1,
			Frequency:          95,
			ExampleSentence:    "Children learn quickly.",
			ExampleTranslation: "Trẻ em học hỏi nhanh chóng.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "practice",
			Phonetic:           "/ˈpræktɪs/",
			Translation:        "thực hành",
			Definition:         "To do something repeatedly to improve",
			PartOfSpeech:       "verb",
			Level:              "intermediate",
			Tags:               "education,improvement,skill",
			Difficulty:         2,
			Frequency:          80,
			ExampleSentence:    "Practice makes perfect.",
			ExampleTranslation: "Luyện tập tạo nên sự hoàn hảo.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "understand",
			Phonetic:           "/ˌʌndərˈstænd/",
			Translation:        "hiểu",
			Definition:         "To know the meaning of something",
			PartOfSpeech:       "verb",
			Level:              "intermediate",
			Tags:               "comprehension,knowledge,thinking",
			Difficulty:         2,
			Frequency:          85,
			ExampleSentence:    "Do you understand this lesson?",
			ExampleTranslation: "Bạn có hiểu bài học này không?",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
		{
			Word:               "remember",
			Phonetic:           "/rɪˈmembər/",
			Translation:        "nhớ",
			Definition:         "To keep something in your mind",
			PartOfSpeech:       "verb",
			Level:              "intermediate",
			Tags:               "memory,thinking,recall",
			Difficulty:         2,
			Frequency:          75,
			ExampleSentence:    "Please remember to do your homework.",
			ExampleTranslation: "Hãy nhớ làm bài tập về nhà.",
			CreatedBy:          1,
			UpdatedBy:          1,
		},
	}

	// Create vocabularies
	for i := range vocabularies {
		// Check if vocabulary already exists
		var existingVocab models.Vocabulary
		if err := db.MasterDB.Where("word = ?", vocabularies[i].Word).First(&existingVocab).Error; err == nil {
			fmt.Printf("⚠️ Vocabulary '%s' already exists, skipping...\n", vocabularies[i].Word)
			vocabularies[i].ID = existingVocab.ID
			continue
		}

		if err := db.MasterDB.Create(&vocabularies[i]).Error; err != nil {
			log.Printf("Failed to create vocabulary %s: %v", vocabularies[i].Word, err)
			continue
		}
		fmt.Printf("✅ Created vocabulary: %s\n", vocabularies[i].Word)
	}

	// Link vocabularies to lesson 4
	fmt.Println("🔗 Linking vocabularies to lesson 4...")
	for i, vocab := range vocabularies {
		// Check if relationship already exists
		var existingLink models.LessonVocabulary
		if err := db.MasterDB.Where("lesson_id = ? AND vocabulary_id = ?", 4, vocab.ID).First(&existingLink).Error; err == nil {
			fmt.Printf("⚠️ Link between lesson 4 and '%s' already exists, skipping...\n", vocab.Word)
			continue
		}

		lessonVocab := models.LessonVocabulary{
			LessonID:     4,
			VocabularyID: vocab.ID,
			SortOrder:    i + 1,
			IsRequired:   true,
		}

		if err := db.MasterDB.Create(&lessonVocab).Error; err != nil {
			log.Printf("Failed to link vocab %s to lesson 4: %v", vocab.Word, err)
			continue
		}
		fmt.Printf("✅ Linked '%s' to lesson 4\n", vocab.Word)
	}

	fmt.Printf("🎉 Lesson 4 flashcard data creation completed!\n")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   - %d vocabularies for lesson 4\n", len(vocabularies))
	fmt.Printf("   - All linked to lesson: %s\n", lesson.Title)
}
