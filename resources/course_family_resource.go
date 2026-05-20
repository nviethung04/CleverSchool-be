package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func CourseFamilyMemberResource(member dto.CourseFamilyMember) *prot.CourseFamilyMember {
	return &prot.CourseFamilyMember{
		Id:          member.ID,
		Name:        member.Name,
		ObjectTitle: member.ObjectTitle,
		ProgramId:   member.ProgramID,
		ProgramName: member.ProgramName,
		SchoolName:  member.SchoolName,
	}
}

func CourseFamilyResource(family *dto.CourseFamily) *prot.CourseFamilyResponse {
	if family == nil {
		return &prot.CourseFamilyResponse{}
	}

	resp := &prot.CourseFamilyResponse{}
	if family.Parent != nil {
		resp.Parent = CourseFamilyMemberResource(*family.Parent)
	}

	for _, child := range family.Children {
		resp.Children = append(resp.Children, CourseFamilyMemberResource(child))
	}

	return resp
}

