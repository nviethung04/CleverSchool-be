package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppleAppSiteAssociationController xử lý endpoint Universal Links cho iOS
type AppleAppSiteAssociationController struct{}

// NewAppleAppSiteAssociationController tạo controller
func NewAppleAppSiteAssociationController() *AppleAppSiteAssociationController {
	return &AppleAppSiteAssociationController{}
}

// appLinksDetail cấu trúc 1 entry trong applinks.details
type appLinksDetail struct {
	AppID string   `json:"appID"`
	Paths []string `json:"paths"`
}

// appleAppSiteAssociationResponse payload cho iOS Universal Links
type appleAppSiteAssociationResponse struct {
	Applinks struct {
		Apps    []interface{}   `json:"apps"`
		Details []appLinksDetail `json:"details"`
	} `json:"applinks"`
}

// GetAppleAppSiteAssociation trả về file apple-app-site-association cho iOS Universal Links.
// Apple request: GET /.well-known/apple-app-site-association
func (ctrl *AppleAppSiteAssociationController) GetAppleAppSiteAssociation(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, appleAppSiteAssociationResponse{
		Applinks: struct {
			Apps    []interface{}   `json:"apps"`
			Details []appLinksDetail `json:"details"`
		}{
			Apps: []interface{}{},
			Details: []appLinksDetail{
				{
					AppID: "R5XX8GYUUL.vn.Clever School.online",
					Paths: []string{"*"},
				},
			},
		},
	})
}
