-- 1) XÓA BẢN GHI SOFT DELETE (chỉ cho các bảng có cột deleted_at)
DELETE FROM exams      WHERE deleted_at IS NOT NULL;
DELETE FROM homeworks  WHERE deleted_at IS NOT NULL;
DELETE FROM lesson_plans WHERE deleted_at IS NOT NULL;
DELETE FROM lessons      WHERE deleted_at IS NOT NULL;
DELETE FROM chapters     WHERE deleted_at IS NOT NULL;
DELETE FROM courses      WHERE deleted_at IS NOT NULL;

-- 2) Xóa chapters không có lessons
DELETE FROM chapters c
WHERE NOT EXISTS (
    SELECT 1 FROM lessons l WHERE l.chapter_id = c.id
);

-- 2b) Xóa EXAMS/HOMEWORKS mồ côi (không có lesson cha)
DELETE FROM exams e
WHERE e.lesson_id IS NULL
   OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = e.lesson_id);

DELETE FROM homeworks h
WHERE h.lesson_id IS NULL
   OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = h.lesson_id);

-- 3) Xóa lessons mồ côi HOẶC lessons không còn liên kết gì (an toàn với FK)
DELETE FROM lessons l
WHERE NOT EXISTS (SELECT 1 FROM chapters c WHERE c.id = l.chapter_id)            -- mồ côi chapter
   OR (
        NOT EXISTS (SELECT 1 FROM lesson_plan_ref_lessons r WHERE r.lesson_id = l.id)  -- không dính lesson_plan
    AND NOT EXISTS (SELECT 1 FROM exams e WHERE e.lesson_id = l.id)                    -- không có exam
    AND NOT EXISTS (SELECT 1 FROM homeworks h WHERE h.lesson_id = l.id)                -- không có homework
   );

-- 4) Xóa lesson_plans không còn liên kết tới parts hoặc ref_lessons
DELETE FROM lesson_plans lp
WHERE NOT EXISTS (SELECT 1 FROM lesson_plan_parts lpp WHERE lpp.lesson_plan_id = lp.id)
  AND NOT EXISTS (SELECT 1 FROM lesson_plan_ref_lessons r WHERE r.lesson_plan_id = lp.id);

-- 5) Xóa lesson_plan_parts không có lesson_plan cha
DELETE FROM lesson_plan_parts lpp
WHERE NOT EXISTS (
    SELECT 1 FROM lesson_plans lp WHERE lp.id = lpp.lesson_plan_id
);

-- 6) Xóa lesson_plan_ref_lessons không có lesson hoặc lesson_plan
DELETE FROM lesson_plan_ref_lessons r
WHERE NOT EXISTS (SELECT 1 FROM lesson_plans lp WHERE lp.id = r.lesson_plan_id)
   OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = r.lesson_id);
