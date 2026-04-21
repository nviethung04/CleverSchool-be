package controllers

import (
	"be-lms/config"
	"be-lms/repositories"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublicMeetingJoinController struct {
	microsoftAttendance services.MeetingAttendanceService
	googleAttendance    services.GoogleMeetingAttendanceService
	microsoftMeeting    repositories.MicrosoftMeetingRepository
	googleMeeting       repositories.GoogleMeetingRepository
	authRepo            repositories.AuthRepository
}

func NewPublicMeetingJoinController(authRepo repositories.AuthRepository) *PublicMeetingJoinController {
	return &PublicMeetingJoinController{
		microsoftAttendance: services.NewMeetingAttendanceService(),
		googleAttendance:    services.NewGoogleMeetingAttendanceService(),
		microsoftMeeting:    repositories.NewMicrosoftMeetingRepository(),
		googleMeeting:       repositories.NewGoogleMeetingRepository(),
		authRepo:            authRepo,
	}
}

func (c *PublicMeetingJoinController) Join(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		utils.Respond(ctx, nil, fmt.Errorf("code is required"), "code is required", http.StatusBadRequest)
		return
	}

	deviceInfo := ctx.GetHeader("User-Agent")
	ipAddress := ctx.ClientIP()

	if userID := ctx.GetInt("currentUserId"); userID > 0 {
		roleId := utils.GetCurrentRoleId(ctx)
		if meeting, _, err := c.microsoftAttendance.JoinByShortCode(code, uint(userID), roleId, deviceInfo, ipAddress); err == nil {
			ctx.Redirect(http.StatusFound, meeting.JoinURL)
			return
		}
		if meeting, _, err := c.googleAttendance.JoinByShortCode(code, uint(userID), roleId, deviceInfo, ipAddress); err == nil {
			ctx.Redirect(http.StatusFound, meeting.MeetURL)
			return
		}
	}

	if meeting, err := c.microsoftMeeting.GetByShortCode(code); err == nil && meeting != nil {
		ctx.Redirect(http.StatusFound, meeting.JoinURL)
		return
	}
	if meeting, err := c.googleMeeting.GetByShortCode(code); err == nil && meeting != nil {
		ctx.Redirect(http.StatusFound, meeting.MeetURL)
		return
	}
	config.Log.Infof("PublicJoin: meeting not found for code %s", code)

	utils.Respond(ctx, nil, fmt.Errorf("meeting not found"), "meeting not found", http.StatusNotFound)
}
