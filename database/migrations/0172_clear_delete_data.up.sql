DO $$
BEGIN
    -- Soft-delete cleanup only when deleted_at exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='exams' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM exams WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='homeworks' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM homeworks WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plans' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM lesson_plans WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lessons' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM lessons WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='chapters' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM chapters WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='courses' AND column_name='deleted_at') THEN
        EXECUTE 'DELETE FROM courses WHERE deleted_at IS NOT NULL';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='chapters')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lessons') THEN
        EXECUTE 'DELETE FROM chapters c WHERE NOT EXISTS (SELECT 1 FROM lessons l WHERE l.chapter_id = c.id)';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='exams')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lessons')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='exams' AND column_name='lesson_id') THEN
        EXECUTE 'DELETE FROM exams e WHERE e.lesson_id IS NULL OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = e.lesson_id)';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homeworks')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lessons')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='homeworks' AND column_name='lesson_id') THEN
        EXECUTE 'DELETE FROM homeworks h WHERE h.lesson_id IS NULL OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = h.lesson_id)';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lessons')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='chapters')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plan_ref_lessons')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='exams')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homeworks')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lessons' AND column_name='chapter_id')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='exams' AND column_name='lesson_id')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='homeworks' AND column_name='lesson_id') THEN
        EXECUTE 'DELETE FROM lessons l
                 WHERE NOT EXISTS (SELECT 1 FROM chapters c WHERE c.id = l.chapter_id)
                    OR (
                        NOT EXISTS (SELECT 1 FROM lesson_plan_ref_lessons r WHERE r.lesson_id = l.id)
                    AND NOT EXISTS (SELECT 1 FROM exams e WHERE e.lesson_id = l.id)
                    AND NOT EXISTS (SELECT 1 FROM homeworks h WHERE h.lesson_id = l.id)
                 )';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plans')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plan_parts')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plan_ref_lessons')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plan_parts' AND column_name='lesson_plan_id')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plan_ref_lessons' AND column_name='lesson_plan_id') THEN
        EXECUTE 'DELETE FROM lesson_plans lp
                 WHERE NOT EXISTS (SELECT 1 FROM lesson_plan_parts lpp WHERE lpp.lesson_plan_id = lp.id)
                   AND NOT EXISTS (SELECT 1 FROM lesson_plan_ref_lessons r WHERE r.lesson_plan_id = lp.id)';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plan_parts')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plans')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plan_parts' AND column_name='lesson_plan_id') THEN
        EXECUTE 'DELETE FROM lesson_plan_parts lpp
                 WHERE NOT EXISTS (SELECT 1 FROM lesson_plans lp WHERE lp.id = lpp.lesson_plan_id)';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plan_ref_lessons')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_plans')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lessons')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plan_ref_lessons' AND column_name='lesson_plan_id')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='lesson_plan_ref_lessons' AND column_name='lesson_id') THEN
        EXECUTE 'DELETE FROM lesson_plan_ref_lessons r
                 WHERE NOT EXISTS (SELECT 1 FROM lesson_plans lp WHERE lp.id = r.lesson_plan_id)
                    OR NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = r.lesson_id)';
    END IF;
END $$;
