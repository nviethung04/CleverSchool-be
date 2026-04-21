CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    number INT NOT NULL,
    name_vn VARCHAR(50),
    name_en VARCHAR(50),
    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

INSERT INTO
    grades (number, name_vn, name_en)
VALUES
    (1, 'Khối 1', 'Grade 1'),
    (2, 'Khối 2', 'Grade 2'),
    (3, 'Khối 3', 'Grade 3'),
    (4, 'Lớp 4', 'Grade 4'),
    (5, 'Khối 5', 'Grade 5');

INSERT INTO
    permissions (
        name,
        permission,
        sort_position,
        "group",
        is_display
    )
VALUES
    (
        'Xem danh sách khối lớp',
        'grades.index',
        21,
        'Khối lớp',
        TRUE
    ),
    (
        'Tạo khối lớp',
        'grades.store',
        21,
        'Khối lớp',
        TRUE
    ),
    (
        'Xóa khối lớp',
        'grades.destroy',
        21,
        'Khối lớp',
        TRUE
    ),
    (
        'Khôi phục khối lớp',
        'grades.restore',
        21,
        'Khối lớp',
        TRUE
    ),
    (
        'Xem chi tiết khối lớp',
        'grades.show',
        21,
        'Khối lớp',
        TRUE
    );
