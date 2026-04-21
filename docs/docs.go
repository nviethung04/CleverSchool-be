package docs

import (
	_ "embed"

	"github.com/swaggo/swag"
)

var docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Token",
            "in": "header",
            "description": "Truyền JWT Token trực tiếp vào header"
        }
    },
	"parameters": {
		"AcceptHeader": {
			"name": "Accept",
			"in": "header",
			"required": true,
			"type": "string",
			"default": "application/json"
		},
		"LocaleHeader": {
			"name": "X-Locale",
			"in": "header",
			"required": false,
			"type": "string",
			"default": "vi"
		}
	},
    "paths": {
        ` +
	authPath + `,` + rolePath + `,` + permissionPath + `,` + userPath + `,` + schoolDashboardPath + `,` +
	schoolPath + `,` + classPath + `,` + subjectPath + `,` + programPath + `,` +
	headingPath + `,` + coursePath + `,` + courseSchedulePath + `,` +
	chapterPath + `,` + lessonPath + `,` + sourceQuestionPath + `,` + questionPath + `,` +
	lessonPlanPath + `,` + lessonPlanPart + `,` + exam + `,` + homework + `,` + assessment + `,` + assessmentCriterion + `,` + assessmentSubcriterion + `,` + assessmentCriteriaGroup + `,` + employeePositionPath + `,` +
	departmentPath + `,` + degreePath + `,` + certificatePath + `,` + tagPath + `,` + skillPath + `,` +
	topicPath + `,` + questionAttributePath + `,` + mediaPath + `,` + saveScore + `,` +
	gradePath + `,` + regionPath + `,` + lessonSchedulePath + `,` +
	weekPath + `,` + studyShiftPath + `,` + dashboardPath + `,` +
	h5pPath + `,` + semesterPath + `,` + holidayPath + `,` + uploadPath + `,` +
	chatPath + `,` + warningPath + `,` + flashcardPath + `,` + doHomework + `,` +
	dashboardTeacherScoringPath + `,` + internalCommandPath + `,` + contest + `,` +
	contestScoring + `,` + dashboardCoursesPath + `,` + dashboardListEntityPath + `,` +
	settingPath + `,` + dashboardTeacherHomeworkListPath + `,` + facultyPath + `,` +
	studyReportPath + `,` + studyReportCriteriaPath + `,` + noticePath + `,` +
	pushFirebasePath + `,` + trainingLevelPath + `,` + teachingPlanPath + `,` +
	`
	},
    "definitions": {
        ` +
	skillProt + `,` + tagProt + `,` + topicProt + `,` + questionAttributeProt + `,` +
	authProt + `,` + roleProt + `,` + permissionProt + `,` + userProt + `,` +
	classProt + `,` + subjectProt + `,` + programProt + `,` +
	headingProt + `,` + courseProt + `,` + courseScheduleProt + `,` +
	chapterProt + `,` + lessonProt + `,` + sourceQuestionProt + `,` + questionProt + `,` +
	lessonPlanProt + `,` + employeePositionProt + `,` + departmentsProt + `,` +
	degreeProt + `,` + certificateProt + `,` + mediaProt + `,` + schoolProt + `,` +
	gradeProt + `,` + regionProt + `,` + lessonScheduleProt + `,` +
	weekProt + `,` + studyShiftProt + `,` + dashboardProt + `,` +
	h5pProt + `,` + semesterProt + `,` + holidayProt + `,` +
	chatProt + `,` + warningProt + `,` + flashcardProt + `,` + examProt + `,` + homeworkProt + `,` + assessmentProt + `,` + assessmentCriterionProt + `,` + assessmentSubcriterionProt + `,` +
	dashboardTeacherScoringProt + `,` + internalCommandProt + `,` + contest + `,` +
	contestScoring + `,` + submitHomeworkProt + `,` + dashboardListEntityProt + `,` +
	settingProt + `,` + dashboardTeacherHomeworkListProt + `,` + facultyProt + `,` +
	studyReportProt + `,` + studyReportCriteriaProt + `,` + noticeProt + `,` +
	pushFirebaseProt + `,` + trainingLevelProt + `,` + teachingPlanProt + `,` +
	`
	}
}`

// SwaggerInfo holds the Swagger specification
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
