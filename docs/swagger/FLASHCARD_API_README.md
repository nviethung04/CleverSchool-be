# 📚 Flashcard API Documentation

## Overview
Comprehensive API documentation for the Flashcard learning system that enables students to learn vocabulary through interactive flashcards with audio, images, and pronunciation assessment.

## 🎯 API Categories

### 1. **Vocabulary Management** (`Flashcard - Vocabulary`)
Manage vocabulary words with multimedia content.

#### Endpoints:
- `GET /api/manage/vocabularies` - List vocabularies with pagination and filtering
- `POST /api/manage/vocabularies` - Create new vocabulary
- `GET /api/manage/vocabularies/{id}` - Get vocabulary details
- `PUT /api/manage/vocabularies/{id}` - Update vocabulary
- `DELETE /api/manage/vocabularies/{id}` - Delete vocabulary

#### Key Features:
- **Multi-modal content**: Word + phonetic + translation + definition + image + audio
- **Example sentences**: With translations and audio
- **Difficulty levels**: 1-5 rating system
- **Tagging system**: For categorization and filtering
- **Metadata**: Part of speech, frequency, level (beginner/intermediate/advanced)

### 2. **Lesson Integration** (`Flashcard - Lesson`)
Link vocabularies to lessons for structured learning.

#### Endpoints:
- `GET /api/manage/lessons/{lessonId}/vocabularies` - Get lesson vocabularies with progress
- `POST /api/manage/lessons/{lessonId}/vocabularies` - Add vocabularies to lesson

#### Key Features:
- **Ordered vocabulary lists**: Custom sort order within lessons
- **Required/Optional marking**: Control which vocabularies are mandatory
- **Progress tracking**: Individual student progress per vocabulary
- **Lesson statistics**: Count of new/learning/known/mastered vocabularies

### 3. **Flashcard Study** (`Flashcard - Study`)
Interactive study sessions with activity tracking.

#### Endpoints:
- `GET /api/manage/lessons/{lessonId}/flashcards` - Get study-ready flashcards
- `POST /api/manage/flashcard-sessions` - Start new study session  
- `PUT /api/manage/flashcard-sessions/{sessionId}` - Update/complete session
- `POST /api/manage/flashcard-sessions/{sessionId}/activities` - Record activity

#### Key Features:
- **Session management**: Track study time and progress
- **Activity logging**: View, flip, mark known/unknown, pronunciation attempts
- **Response time tracking**: Measure learning efficiency
- **Shuffle support**: Randomize card order for better learning

### 4. **Progress Tracking** (`Flashcard - Progress`)
Monitor learning progress and mastery levels.

#### Endpoints:
- `PUT /api/manage/vocabulary-progress/{vocabularyId}` - Update progress status

#### Progress States:
- **New**: Never studied
- **Learning**: Actively studying
- **Known**: Student marked as known
- **Mastered**: Fully learned with high accuracy

### 5. **Pronunciation Assessment** (`Flashcard - Pronunciation`)
AI-powered pronunciation scoring and feedback.

#### Endpoints:
- `POST /api/manage/vocabulary/{vocabularyId}/pronunciation` - Submit audio for assessment

#### Features:
- **Audio submission**: Base64 encoded audio data
- **Score feedback**: 0-100 accuracy rating
- **Mistake identification**: Phoneme-level error detection
- **Improvement suggestions**: AI-generated feedback

### 6. **Analytics & Statistics** (`Flashcard - Analytics`)
Comprehensive learning analytics and reporting.

#### Endpoints:
- `GET /api/manage/students/{studentId}/flashcard-stats` - Get student statistics

#### Analytics Include:
- **Progress overview**: Total/new/learning/known/mastered counts
- **Study habits**: Total sessions, study time, streaks
- **Performance metrics**: Average scores, best scores
- **Pronunciation progress**: Average and best pronunciation scores
- **Time-based analysis**: Daily/weekly/monthly progress

## 🔐 Authentication
All endpoints require JWT Bearer token authentication:
```
Authorization: Bearer <jwt_token>
```

## 📊 Response Format
All API responses follow a consistent format:
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { /* Response data */ },
  "pagination": { /* Pagination info for lists */ }
}
```

## 🎨 Request Examples

### Create Vocabulary
```json
POST /api/manage/vocabularies
{
  "word": "hello",
  "phonetic": "/həˈloʊ/",
  "translation": "xin chào",
  "definition": "A greeting used when meeting someone",
  "part_of_speech": "interjection",
  "level": "beginner",
  "difficulty": 2,
  "word_audio_id": 123,
  "image_id": 456,
  "example_sentence": "Hello, how are you today?",
  "example_translation": "Xin chào, hôm nay bạn khỏe không?",
  "example_audio_id": 789
}
```

### Start Study Session
```json
POST /api/manage/flashcard-sessions
{
  "lesson_id": 42
}
```

### Record Activity
```json
POST /api/manage/flashcard-sessions/123/activities?vocabularyId=456
{
  "activity_type": "mark_known",
  "response": "known",
  "response_time": 2500
}
```

### Pronunciation Assessment
```json
POST /api/manage/vocabulary/456/pronunciation
{
  "vocabulary_id": 456,
  "audio_data": "data:audio/wav;base64,UklGRiQAAABXQVZF...",
  "word_type": "word"
}
```

## 🚀 Getting Started

1. **Authentication**: Obtain JWT token from `/api/login`
2. **Create Vocabularies**: Add words with multimedia content
3. **Link to Lessons**: Associate vocabularies with lesson structure
4. **Study Sessions**: Students start flashcard sessions
5. **Track Progress**: Monitor learning advancement
6. **Assess Pronunciation**: Get AI feedback on speaking

## 📈 Advanced Features

### Adaptive Learning
- Skip "Remember" and "Super Easy" exercises for known vocabularies
- Full exercise set for unknown vocabularies
- Dynamic difficulty adjustment based on performance

### Gamification
- Study streaks tracking
- Session scoring system
- Progress badges and achievements
- Leaderboards and competitions

### Multi-Platform Support
- Web responsive design
- Mobile app optimization
- Offline mode capability
- Cross-device synchronization

---

For detailed request/response schemas, parameter descriptions, and error codes, visit the interactive Swagger UI at `/swagger/index.html` when the server is running.
