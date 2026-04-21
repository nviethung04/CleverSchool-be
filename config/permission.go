package config

type PermissionGroup struct {
	Group        string
	SortPosition int
	Names        []string
	Actions      []string
}

func GetPermissions() map[string]PermissionGroup {
	return map[string]PermissionGroup{
		"users": {
			Group:        "Tài khoản",
			SortPosition: 1,
			Names:        []string{"Xem danh sách tài khoản", "Tạo tài khoản", "Sửa tài khoản", "Xoá tài khoản", "Xem chi tiết tài khoản", "Export tài khoản", "Import tài khoản"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "export", "import"},
		},
		"roles": {
			Group:        "Vai trò",
			SortPosition: 2,
			Names:        []string{"Xem danh sách vai trò", "Tạo vai trò", "Sửa vai trò", "Xoá vai trò", "Xem chi tiết vai trò"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"permissions": {
			Group:        "Phân quyền",
			SortPosition: 3,
			Names:        []string{"Sửa phân quyền", "Xem phân quyền"},
			Actions:      []string{"update", "show"},
		},
		"schools": {
			Group:        "Trường học",
			SortPosition: 4,
			Names:        []string{"Xem danh sách trường học", "Tạo trường học", "Sửa trường học", "Xoá trường học", "Xem chi tiết trường học", "Khôi phục trường học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"classes": {
			Group:        "Lớp học",
			SortPosition: 5,
			Names:        []string{"Xem danh sách lớp học", "Tạo lớp học", "Sửa lớp học", "Xoá lớp học", "Xem chi tiết lớp học", "Khôi phục lớp học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"subjects": {
			Group:        "Môn học",
			SortPosition: 6,
			Names:        []string{"Xem danh sách môn học", "Tạo môn học", "Sửa môn học", "Xoá môn học", "Xem chi tiết môn học", "Khôi phục môn học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"courses": {
			Group:        "Khóa học",
			SortPosition: 7,
			Names:        []string{"Xem danh sách khoá học", "Tạo khoá học", "Sửa khoá học", "Xoá khoá học", "Xem chi tiết khoá học", "Khôi phục khoá học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lessons": {
			Group:        "Bài học",
			SortPosition: 8,
			Names:        []string{"Xem danh sách bài học", "Tạo bài học", "Sửa bài học", "Xoá bài học", "Xem chi tiết bài học", "Khôi phục tiết học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lesson-plans": {
			Group:        "Giáo án",
			SortPosition: 9,
			Names:        []string{"Xem danh sách giáo án", "Tạo giáo án", "Sửa giáo án", "Xoá giáo án", "Xem chi tiết giáo án"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"lesson-plan-parts": {
			Group:        "Kế hoạch bài học",
			SortPosition: 10,
			Names:        []string{"Xem danh sách kế hoạch bài học", "Tạo kế hoạch bài học", "Sửa kế hoạch bài học", "Xoá kế hoạch bài học", "Xem chi tiết kế hoạch bài học"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"chapters": {
			Group:        "Chương học",
			SortPosition: 11,
			Names:        []string{"Xem danh sách chương học", "Tạo chương học", "Sửa chương học", "Xoá chương học", "Xem chi tiết chương học", "Khôi phục chương học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"source-questions": {
			Group:        "Nguồn câu hỏi",
			SortPosition: 12,
			Names:        []string{"Xem danh sách nguồn câu hỏi", "Tạo nguồn câu hỏi", "Sửa nguồn câu hỏi", "Xoá nguồn câu hỏi", "Xem chi tiết nguồn câu hỏi", "Khôi phục nguồn câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"questions": {
			Group:        "Câu hỏi",
			SortPosition: 13,
			Names:        []string{"Xem danh sách câu hỏi", "Tạo câu hỏi", "Sửa câu hỏi", "Xoá câu hỏi", "Xem chi tiết câu hỏi", "Khôi phục câu hỏi", "Export câu hỏi", "Import câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore", "export", "import"},
		},
		"exams": {
			Group:        "Bài kiểm tra",
			SortPosition: 14,
			Names:        []string{"Xem danh sách bài kiểm tra", "Tạo bài kiểm tra", "Sửa bài kiểm tra", "Xoá bài kiểm tra", "Xem chi tiết bài kiểm tra"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"homeworks": {
			Group:        "Bài tập về nhà",
			SortPosition: 15,
			Names:        []string{"Xem danh sách bài tập về nhà", "Tạo bài tập về nhà", "Sửa bài tập về nhà", "Xoá bài tập về nhà", "Xem chi tiết bài tập về nhà"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"assessments": {
			Group:        "Đánh giá",
			SortPosition: 15,
			Names:        []string{"Xem danh sách đánh giá", "Tạo đánh giá", "Sửa đánh giá", "Xoá đánh giá", "Xem chi tiết đánh giá"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"assessment-criteria": {
			Group:        "Tiêu chí đánh giá",
			SortPosition: 15,
			Names:        []string{"Xem danh sách tiêu chí đánh giá", "Tạo tiêu chí đánh giá", "Sửa tiêu chí đánh giá", "Xoá tiêu chí đánh giá", "Xem chi tiết tiêu chí đánh giá"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"assessment-subcriteria": {
			Group:        "Tiêu chí con",
			SortPosition: 15,
			Names:        []string{"Xem danh sách tiêu chí con", "Tạo tiêu chí con", "Sửa tiêu chí con", "Xoá tiêu chí con", "Xem chi tiết tiêu chí con"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"assessment-criteria-groups": {
			Group:        "Nhóm tiêu chí đánh giá",
			SortPosition: 15,
			Names:        []string{"Xem danh sách nhóm tiêu chí", "Tạo nhóm tiêu chí", "Sửa nhóm tiêu chí", "Xoá nhóm tiêu chí", "Xem chi tiết nhóm tiêu chí"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"departments": {
			Group:        "Phòng ban",
			SortPosition: 15,
			Names:        []string{"Xem danh sách phòng ban", "Tạo phòng ban", "Sửa phòng ban", "Xoá phòng ban", "Xem chi tiết phòng ban", "Khôi phục phòng ban"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"employee-positions": {
			Group:        "Chức vụ",
			SortPosition: 15,
			Names:        []string{"Xem danh sách chức vụ", "Tạo chức vụ", "Sửa chức vụ", "Xoá chức vụ", "Xem chi tiết chức vụ", "Khôi phục chức vụ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"degrees": {
			Group:        "Bằng cấp",
			SortPosition: 15,
			Names:        []string{"Xem danh sách bằng cấp", "Tạo bằng cấp", "Sửa bằng cấp", "Xoá bằng cấp", "Xem chi tiết bằng cấp", "Khôi phục bằng cấp"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"certificates": {
			Group:        "Chứng chỉ",
			SortPosition: 15,
			Names:        []string{"Xem danh sách chứng chỉ", "Tạo chứng chỉ", "Sửa chứng chỉ", "Xoá chứng chỉ", "Xem chứng chỉ", "Khôi phục chứng chỉ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"question-attributes": {
			Group:        "Thuộc tính câu hỏi",
			SortPosition: 16,
			Names:        []string{"Xem danh sách thuộc tính câu hỏi", "Tạo thuộc tính câu hỏi", "Sửa thuộc tính câu hỏi", "Xoá thuộc tính câu hỏi", "Xem chi tiết thuộc tính câu hỏi", "Khôi phục thuộc tính câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"tags": {
			Group:        "Thẻ",
			SortPosition: 17,
			Names:        []string{"Xem danh sách thẻ", "Tạo thẻ", "Sửa thẻ", "Xoá thẻ", "Xem chi tiết thẻ", "Khôi phục thẻ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"topics": {
			Group:        "Chủ đề",
			SortPosition: 18,
			Names:        []string{"Xem danh sách chủ đề", "Tạo chủ đề", "Sửa chủ đề", "Xoá chủ đề", "Xem chi tiết chủ đề", "Khôi phục chủ đề"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"skills": {
			Group:        "Kỹ năng",
			SortPosition: 19,
			Names:        []string{"Xem danh sách kỹ năng", "Tạo kỹ năng", "Sửa kỹ năng", "Xoá kỹ năng", "Xem chi tiết kỹ năng", "Khôi phục kỹ năng"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"study-shifts": {
			Group:        "Ca học",
			SortPosition: 20,
			Names:        []string{"Xem danh sách ca học", "Tạo ca học", "Sửa ca học", "Xoá ca học", "Xem chi tiết ca học", "Khôi phục ca học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
	}
}

func GetTeacherPermissions() map[string]PermissionGroup {
	return map[string]PermissionGroup{
		"users": {
			Group:        "Tài khoản",
			SortPosition: 1,
			Names:        []string{"Tạo tài khoản", "Sửa tài khoản", "Xoá tài khoản", "Xem chi tiết tài khoản"},
			Actions:      []string{"store", "update", "destroy", "show"},
		},
		"schools": {
			Group:        "Trường học",
			SortPosition: 4,
			Names:        []string{"Xem danh sách trường học", "Tạo trường học", "Sửa trường học", "Xoá trường học", "Xem chi tiết trường học", "Khôi phục trường học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"classes": {
			Group:        "Lớp học",
			SortPosition: 5,
			Names:        []string{"Xem danh sách lớp học", "Tạo lớp học", "Sửa lớp học", "Xoá lớp học", "Xem chi tiết lớp học", "Khôi phục lớp học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"subjects": {
			Group:        "Môn học",
			SortPosition: 6,
			Names:        []string{"Xem danh sách môn học", "Tạo môn học", "Sửa môn học", "Xoá môn học", "Xem chi tiết môn học", "Khôi phục môn học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"courses": {
			Group:        "Khóa học",
			SortPosition: 7,
			Names:        []string{"Xem danh sách khoá học", "Tạo khoá học", "Sửa khoá học", "Xoá khoá học", "Xem chi tiết khoá học", "Khôi phục khoá học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lessons": {
			Group:        "Bài học",
			SortPosition: 8,
			Names:        []string{"Xem danh sách bài học", "Tạo bài học", "Sửa bài học", "Xoá bài học", "Xem chi tiết bài học", "Khôi phục tiết học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lesson-plans": {
			Group:        "Giáo án",
			SortPosition: 9,
			Names:        []string{"Xem danh sách giáo án", "Tạo giáo án", "Sửa giáo án", "Xoá giáo án", "Xem chi tiết giáo án"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"lesson-plan-parts": {
			Group:        "Kế hoạch bài học",
			SortPosition: 10,
			Names:        []string{"Xem danh sách kế hoạch bài học", "Tạo kế hoạch bài học", "Sửa kế hoạch bài học", "Xoá kế hoạch bài học", "Xem chi tiết kế hoạch bài học"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"chapters": {
			Group:        "Chương học",
			SortPosition: 11,
			Names:        []string{"Xem danh sách chương học", "Tạo chương học", "Sửa chương học", "Xoá chương học", "Xem chi tiết chương học", "Khôi phục chương học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"source-questions": {
			Group:        "Nguồn câu hỏi",
			SortPosition: 12,
			Names:        []string{"Xem danh sách nguồn câu hỏi", "Tạo nguồn câu hỏi", "Sửa nguồn câu hỏi", "Xoá nguồn câu hỏi", "Xem chi tiết nguồn câu hỏi", "Khôi phục nguồn câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"questions": {
			Group:        "Câu hỏi",
			SortPosition: 13,
			Names:        []string{"Xem danh sách câu hỏi", "Tạo câu hỏi", "Sửa câu hỏi", "Xoá câu hỏi", "Xem chi tiết câu hỏi", "Khôi phục câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"exams": {
			Group:        "Bài kiểm tra",
			SortPosition: 14,
			Names:        []string{"Xem danh sách bài kiểm tra", "Tạo bài kiểm tra", "Sửa bài kiểm tra", "Xoá bài kiểm tra", "Xem chi tiết bài kiểm tra"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"homeworks": {
			Group:        "Bài tập về nhà",
			SortPosition: 15,
			Names:        []string{"Xem danh sách bài tập về nhà", "Tạo bài tập về nhà", "Sửa bài tập về nhà", "Xoá bài tập về nhà", "Xem chi tiết bài tập về nhà"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"departments": {
			Group:        "Phòng ban",
			SortPosition: 15,
			Names:        []string{"Xem danh sách phòng ban", "Tạo phòng ban", "Sửa phòng ban", "Xoá phòng ban", "Xem chi tiết phòng ban", "Khôi phục phòng ban"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"employee-positions": {
			Group:        "Chức vụ",
			SortPosition: 15,
			Names:        []string{"Xem danh sách chức vụ", "Tạo chức vụ", "Sửa chức vụ", "Xoá chức vụ", "Xem chi tiết chức vụ", "Khôi phục chức vụ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"degrees": {
			Group:        "Bằng cấp",
			SortPosition: 15,
			Names:        []string{"Xem danh sách bằng cấp", "Tạo bằng cấp", "Sửa bằng cấp", "Xoá bằng cấp", "Xem chi tiết bằng cấp", "Khôi phục bằng cấp"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"certificates": {
			Group:        "Chứng chỉ",
			SortPosition: 15,
			Names:        []string{"Xem danh sách chứng chỉ", "Tạo chứng chỉ", "Sửa chứng chỉ", "Xoá chứng chỉ", "Xem chứng chỉ", "Khôi phục chứng chỉ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"question-attributes": {
			Group:        "Thuộc tính câu hỏi",
			SortPosition: 16,
			Names:        []string{"Xem danh sách thuộc tính câu hỏi", "Tạo thuộc tính câu hỏi", "Sửa thuộc tính câu hỏi", "Xoá thuộc tính câu hỏi", "Xem chi tiết thuộc tính câu hỏi", "Khôi phục thuộc tính câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"tags": {
			Group:        "Thẻ",
			SortPosition: 17,
			Names:        []string{"Xem danh sách thẻ", "Tạo thẻ", "Sửa thẻ", "Xoá thẻ", "Xem chi tiết thẻ", "Khôi phục thẻ"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},

		"topics": {
			Group:        "Chủ đề",
			SortPosition: 18,
			Names:        []string{"Xem danh sách chủ đề", "Tạo chủ đề", "Sửa chủ đề", "Xoá chủ đề", "Xem chi tiết chủ đề", "Khôi phục chủ đề"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"skills": {
			Group:        "Kỹ năng",
			SortPosition: 19,
			Names:        []string{"Xem danh sách kỹ năng", "Tạo kỹ năng", "Sửa kỹ năng", "Xoá kỹ năng", "Xem chi tiết kỹ năng", "Khôi phục kỹ năng"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"study-shifts": {
			Group:        "Ca học",
			SortPosition: 20,
			Names:        []string{"Xem danh sách ca học", "Tạo ca học", "Sửa ca học", "Xoá ca học", "Xem chi tiết ca học", "Khôi phục ca học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
	}
}

func GetStudentPermissions() map[string]PermissionGroup {
	return map[string]PermissionGroup{
		"users": {
			Group:        "Tài khoản",
			SortPosition: 1,
			Names:        []string{"Tạo tài khoản", "Sửa tài khoản", "Xoá tài khoản", "Xem chi tiết tài khoản"},
			Actions:      []string{"store", "update", "destroy", "show"},
		},
		"schools": {
			Group:        "Trường học",
			SortPosition: 4,
			Names:        []string{"Xem danh sách trường học", "Tạo trường học", "Sửa trường học", "Xoá trường học", "Xem chi tiết trường học", "Khôi phục trường học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"classes": {
			Group:        "Lớp học",
			SortPosition: 5,
			Names:        []string{"Xem danh sách lớp học", "Tạo lớp học", "Sửa lớp học", "Xoá lớp học", "Xem chi tiết lớp học", "Khôi phục lớp học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"subjects": {
			Group:        "Môn học",
			SortPosition: 6,
			Names:        []string{"Xem danh sách môn học", "Tạo môn học", "Sửa môn học", "Xoá môn học", "Xem chi tiết môn học", "Khôi phục môn học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"courses": {
			Group:        "Khóa học",
			SortPosition: 7,
			Names:        []string{"Xem danh sách khoá học", "Tạo khoá học", "Sửa khoá học", "Xoá khoá học", "Xem chi tiết khoá học", "Khôi phục khoá học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lessons": {
			Group:        "Bài học",
			SortPosition: 8,
			Names:        []string{"Xem danh sách bài học", "Tạo bài học", "Sửa bài học", "Xoá bài học", "Xem chi tiết bài học", "Khôi phục tiết học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"lesson-plans": {
			Group:        "Giáo án",
			SortPosition: 9,
			Names:        []string{"Xem danh sách giáo án", "Tạo giáo án", "Sửa giáo án", "Xoá giáo án", "Xem chi tiết giáo án"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"lesson-plan-parts": {
			Group:        "Kế hoạch bài học",
			SortPosition: 10,
			Names:        []string{"Xem danh sách kế hoạch bài học", "Tạo kế hoạch bài học", "Sửa kế hoạch bài học", "Xoá kế hoạch bài học", "Xem chi tiết kế hoạch bài học"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"chapters": {
			Group:        "Chương học",
			SortPosition: 11,
			Names:        []string{"Xem danh sách chương học", "Tạo chương học", "Sửa chương học", "Xoá chương học", "Xem chi tiết chương học", "Khôi phục chương học"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"source-questions": {
			Group:        "Nguồn câu hỏi",
			SortPosition: 12,
			Names:        []string{"Xem danh sách nguồn câu hỏi", "Tạo nguồn câu hỏi", "Sửa nguồn câu hỏi", "Xoá nguồn câu hỏi", "Xem chi tiết nguồn câu hỏi", "Khôi phục nguồn câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"questions": {
			Group:        "Câu hỏi",
			SortPosition: 13,
			Names:        []string{"Xem danh sách câu hỏi", "Tạo câu hỏi", "Sửa câu hỏi", "Xoá câu hỏi", "Xem chi tiết câu hỏi", "Khôi phục câu hỏi"},
			Actions:      []string{"index", "store", "update", "destroy", "show", "restore"},
		},
		"exams": {
			Group:        "Bài kiểm tra",
			SortPosition: 14,
			Names:        []string{"Xem danh sách bài kiểm tra", "Tạo bài kiểm tra", "Sửa bài kiểm tra", "Xoá bài kiểm tra", "Xem chi tiết bài kiểm tra"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
		"homeworks": {
			Group:        "Bài tập về nhà",
			SortPosition: 15,
			Names:        []string{"Xem danh sách bài tập về nhà", "Tạo bài tập về nhà", "Sửa bài tập về nhà", "Xoá bài tập về nhà", "Xem chi tiết bài tập về nhà"},
			Actions:      []string{"index", "store", "update", "destroy", "show"},
		},
	}
}
