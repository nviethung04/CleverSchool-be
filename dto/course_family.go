package dto

type CourseFamilyMember struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ObjectTitle   string `json:"object_title"`
	ProgramID     int64  `json:"program_id"`
	ProgramName   string `json:"program_name"`
	SchoolName    string `json:"school_name"`
}

type CourseFamily struct {
	Parent   *CourseFamilyMember   `json:"parent"`
	Children []CourseFamilyMember `json:"children"`
}

