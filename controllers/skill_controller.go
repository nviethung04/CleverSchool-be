package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type SkillController struct {
	*GenericController[models.Skill, prot.Skill, *prot.SkillRequest]
}

func NewSkillController(service services.SkillService) *SkillController {
	skillResource := resources.NewSkillResource()
	skillResourceAdapter := NewSkillResourceAdapter(skillResource)

	genericController := NewGenericController(
		service,
		skillResourceAdapter,
		func() *prot.SkillRequest {
			return &prot.SkillRequest{}
		},
		func(skills []*prot.Skill, totalCount uint64) interface{} {
			return &prot.SkillsResponse{
				Skills:       skills,
				TotalCount: totalCount,
			}
		},
	)

	return &SkillController{
		GenericController: genericController,
	}
}

