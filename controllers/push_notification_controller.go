package controllers

import (
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PushNotificationController struct {
	pushService *services.PushNotificationService
}

func NewPushNotificationController() *PushNotificationController {
	return &PushNotificationController{
		pushService: services.NewPushNotificationService(),
	}
}

func (ctrl *PushNotificationController) RegisterDevice(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	req, err, _ := utils.GetBody[*prot.RegisterDeviceRequest](c, func() *prot.RegisterDeviceRequest {
		return &prot.RegisterDeviceRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	deviceToken := req.DeviceToken

	platform := c.GetHeader("platform")
	appVersion := c.GetHeader("app_version")

	deviceType := detectDeviceType(platform)

	if deviceType == "" {
		utils.Respond(c, nil, fmt.Errorf("Invalid platform header. Must be 'iOS', 'Android', or 'Web'"), "Invalid platform header. Must be 'iOS', 'Android', or 'Web'", http.StatusBadRequest)
		return
	}

	err = ctrl.pushService.RegisterDevice(
		int64(userID),
		deviceToken,
		deviceType,
		platform,
		appVersion,
	)

	if err != nil {
		utils.Respond(c, nil, err, "Failed to register device", http.StatusBadRequest)
		return
	}

	utils.Respond(c, &prot.RegisterDeviceResponse{Message: "Device registered successfully"}, nil, "")
}

func (ctrl *PushNotificationController) UnregisterDevice(c *gin.Context) {
	req, err, _ := utils.GetBody[*prot.UnregisterDeviceRequest](c, func() *prot.UnregisterDeviceRequest {
		return &prot.UnregisterDeviceRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	deviceToken := req.DeviceToken

	err = ctrl.pushService.UnregisterDevice(deviceToken)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to unregister device", http.StatusBadRequest)
		return
	}

	utils.Respond(c, &prot.UnregisterDeviceResponse{Message: "Device unregistered successfully"}, nil, "")
}

func (ctrl *PushNotificationController) GetMyDevices(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	devices, err := ctrl.pushService.GetUserDevices(int64(userID))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to get devices", http.StatusBadRequest)
		return
	}

	var deviceFormats []*prot.UserDevice

	for _, device := range devices {
		deviceFormats = append(deviceFormats, &prot.UserDevice{
			Id:          device.ID,
			UserId:      device.UserID,
			DeviceToken: device.DeviceToken,
			DeviceType:  device.DeviceType,
			DeviceName:  device.DeviceName,
			Platform:    device.Platform,
			AppVersion:  device.AppVersion,
			OsVersion:   device.OsVersion,
			IsActive:    device.IsActive,
			CreatedAt:   device.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   device.UpdatedAt.Format("2006-01-02 15:04:05"),
			LastUsedAt:  device.LastUsedAt.Format("2006-01-02 15:04:05"),
		})
	}

	utils.Respond(c, &prot.MyDeviceResponse{
		Devices: deviceFormats,
		Total:   int32(len(devices)),
	}, nil, "")
}

func (ctrl *PushNotificationController) SendTestNotification(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	req, err, _ := utils.GetBody[*prot.TestNotificationRequest](c, func() *prot.TestNotificationRequest {
		return &prot.TestNotificationRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	data := make(map[string]string)
	if req.Data != nil {
		for k, v := range req.Data {
			data[k] = v
		}
	} else {
		data["type"] = "test"
	}

	err = ctrl.pushService.SendToUser(int64(userID), req.Title, req.Body, data)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to send test notification", http.StatusBadRequest)
		return
	}

	utils.Respond(c, &prot.TestNotificationResponse{Message: "Test notification sent"}, nil, "")
}

func (ctrl *PushNotificationController) SendNoticeWithPush(c *gin.Context) {
	noticeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid notice ID")
		return
	}

	req, err, _ := utils.GetBody[*prot.SendNoticePushRequest](c, func() *prot.SendNoticePushRequest {
		return &prot.SendNoticePushRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}
	validTypes := map[string]bool{"system": true, "school": true, "class": true, "course": true, "user": true}
	if !validTypes[req.Type] {
		utils.Respond(c, nil, fmt.Errorf("invalid type"), "Type must be 'system', 'school', 'class', 'course', or 'user'")
		return
	}

	notice, err := ctrl.pushService.GetNoticeByID(noticeID)
	if err != nil {
		log.Printf("❌ Failed to get notice: %v\n", err)
		utils.Respond(c, nil, err, "Notice not found", http.StatusNotFound)
		return
	}

	err = ctrl.pushService.SendNoticeToTarget(notice, req.Type, req.Value)
	if err != nil {
		log.Printf("❌ SendNoticeToTarget failed: %v\n", err)
		utils.Respond(c, nil, err, "Failed to send push notification")
		return
	}

	utils.Respond(c, &prot.SendNoticeWithPushResponse{
		Message:  "Push notification sent successfully",
		NoticeId: noticeID,
		Target:   fmt.Sprintf("%s:%d", req.Type, req.Value),
	}, nil, "")
}

func detectDeviceType(platform string) string {
	if platform == "" {
		return ""
	}

	if len(platform) >= 3 {
		prefix := platform[:3]
		if prefix == "iOS" || prefix == "ios" {
			return models.DeviceTypeIOS
		}
	}

	if len(platform) >= 7 {
		prefix := platform[:7]
		if prefix == "Android" || prefix == "android" {
			return models.DeviceTypeAndroid
		}
	}

	if platform == "Web" || platform == "web" {
		return models.DeviceTypeWeb
	}

	return ""
}
