-- Best-effort rollback: remove seeded rows only (by known names).

DELETE FROM question_attributes
WHERE name IN ('Vocabulary', 'Structure', 'Phonics', 'Easy', 'Medium', 'Hard',
               'Nhận diện', 'Ghi nhớ', 'Vận dụng', 'Vận dụng nâng cao/sáng tạo');

DELETE FROM question_attributes
WHERE name IN ('Kỹ năng', 'Mức độ', 'Mức độ nhận thức')
  AND COALESCE(parent_id, 0) = 0;
