-- Create unified dashboard overview function for single query approach
CREATE OR REPLACE FUNCTION get_dashboard_overview(
    p_end_time TIMESTAMP,
    p_start_previous_time TIMESTAMP,
    p_end_previous_time TIMESTAMP,
    p_end_previous_time_for_student_teacher TIMESTAMP,
    p_start_time TIMESTAMP,
    p_province_code VARCHAR DEFAULT NULL,
    p_ward_code VARCHAR DEFAULT NULL,
    p_course_id BIGINT DEFAULT NULL,
    p_school_id BIGINT DEFAULT NULL
)
RETURNS TABLE (
    school_total_count BIGINT,
    school_current_count BIGINT,
    school_previous_count BIGINT,
    user_total_count BIGINT,
    user_current_count BIGINT,
    user_previous_count BIGINT,
    student_total_count BIGINT,
    student_current_count BIGINT,
    student_previous_count BIGINT,
    teacher_total_count BIGINT,
    teacher_current_count BIGINT,
    teacher_previous_count BIGINT,
    activity_total_count BIGINT,
    activity_current_count BIGINT,
    activity_previous_count BIGINT,
    activity_new_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH params AS (
        -- determine current range and previous range fallback logic
        SELECT
            -- current range: prefer provided p_start_time/p_end_time, otherwise use calendar month of p_end_time
            COALESCE(p_start_time, date_trunc('month', p_end_time)) AS curr_start,
            COALESCE(p_end_time, (date_trunc('month', p_end_time) + INTERVAL '1 month') - INTERVAL '1 second') AS curr_end,
            -- previous range: prefer provided p_start_previous_time/p_end_previous_time, otherwise previous calendar month
            COALESCE(p_start_previous_time, date_trunc('month', p_end_time) - INTERVAL '1 month') AS prev_start,
            COALESCE(p_end_previous_time, date_trunc('month', p_end_time) - INTERVAL '1 second') AS prev_end
    )
    SELECT
        -- School counts with filters
        -- NOTE: school_total_count is cumulative up to curr_end (i.e. created_at <= curr_end)
        (SELECT COUNT(*) FROM schools s
         LEFT JOIN wards w ON w.code = s.ward_code
         WHERE s.deleted_at IS NULL
           AND s.created_at <= (SELECT curr_end FROM params)
           AND (p_school_id IS NULL OR p_school_id = 0 OR s.id = p_school_id)
           AND (p_ward_code IS NULL OR s.ward_code = p_ward_code)
           AND (p_province_code IS NULL OR w.province_code = p_province_code)
        )::BIGINT,
        -- current schools (use curr_start/curr_end from params)
        (SELECT COUNT(*) FROM schools s
         LEFT JOIN wards w ON w.code = s.ward_code
         WHERE s.deleted_at IS NULL
           AND s.created_at >= (SELECT curr_start FROM params)
           AND s.created_at <= (SELECT curr_end FROM params)
           AND (p_school_id IS NULL OR p_school_id = 0 OR s.id = p_school_id)
           AND (p_ward_code IS NULL OR s.ward_code = p_ward_code)
           AND (p_province_code IS NULL OR w.province_code = p_province_code)
        )::BIGINT,
        -- previous schools (use prev_start/prev_end from params)
        (SELECT COUNT(*) FROM schools s
         LEFT JOIN wards w ON w.code = s.ward_code
         WHERE s.deleted_at IS NULL
           AND s.created_at >= (SELECT prev_start FROM params)
           AND s.created_at <= (SELECT prev_end FROM params)
           AND (p_school_id IS NULL OR p_school_id = 0 OR s.id = p_school_id)
           AND (p_ward_code IS NULL OR s.ward_code = p_ward_code)
           AND (p_province_code IS NULL OR w.province_code = p_province_code)
        )::BIGINT,

        -- User counts with filters
        (SELECT COUNT(*) FROM users u
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at <= p_end_time
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at >= p_start_previous_time AND u.created_at <= p_end_previous_time
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,

        -- Student counts with filters
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at <= p_end_time
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at <= p_end_previous_time_for_student_teacher
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,

        -- Teacher counts with filters
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 2
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 2
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at <= p_end_time
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,
        (SELECT COUNT(*) FROM users u
         JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 2
         LEFT JOIN user_address ua ON ua.user_id = u.id
         WHERE u.deleted_at IS NULL AND u.created_at <= p_end_previous_time_for_student_teacher
         AND (p_school_id IS NULL OR p_school_id = 0 OR u.school_id = p_school_id)
         AND (p_ward_code IS NULL OR ua.ward_code = p_ward_code)
         AND (p_province_code IS NULL OR ua.province_code = p_province_code)
         AND (p_course_id IS NULL OR p_course_id = 0 OR (p_course_id > 0 AND EXISTS (
             SELECT 1 FROM user_courses uc WHERE uc.user_id = u.id AND uc.course_id = p_course_id
         ))))::BIGINT,

        -- Activity counts (will be calculated separately in Go code)
        0::BIGINT, -- activity_total_count
        0::BIGINT, -- activity_current_count
        0::BIGINT, -- activity_previous_count
        0::BIGINT  -- activity_new_count
    ;
END;
$$ LANGUAGE plpgsql;
