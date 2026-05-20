package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	DiskStorage = "public"
	PathStorage = "h5p"
)

type H5pController struct {
	svc services.H5pService
}

func NewH5pController(svc services.H5pService) *H5pController {
	return &H5pController{svc: svc}
}

func (htl *H5pController) ListContent(c *gin.Context) {
	contents, rows, err := htl.svc.ListContent(c)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	h5pPtrs := make([]*models.H5pContent, len(contents))
	for i := range contents {
		h5pPtrs[i] = &contents[i]
	}


	resource := resources.NewH5PResource()
	formatted := resource.FormatH5Ps(h5pPtrs)

	response := &prot.H5PsResponse{
		TotalCount:  uint64(rows),
		Contents: formatted,
	}

	utils.Respond(c, response, err, "")
}

func (htl *H5pController) ShowContent(c *gin.Context) {
	idStr := c.Param("id")
	content, err := htl.svc.ShowContent(c, idStr)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resource := resources.NewH5PResource()
	formatted := resource.FormatH5P(content)

	utils.Respond(c, formatted, err, "")
}

func (htl *H5pController) UpdateContent(c *gin.Context) {
	idStr := c.Param("id")
	updated, err := htl.svc.UpdateContent(c, idStr)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	utils.Respond(c, updated, err, "")
}

func (htl *H5pController) DeleteContent(c *gin.Context) {
	idStr := c.Param("id")
	err := htl.svc.DeleteContent(c, idStr)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (htl *H5pController) StoreScore(c *gin.Context) {
	err := htl.svc.StoreScore(c)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (htl *H5pController) StoreContentUserData(c *gin.Context) {
	err := htl.svc.StoreContentUserData(c)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (htl *H5pController) GetContentUserData(c *gin.Context) {

	data := htl.svc.GetContentUserData(c)

	c.JSON(http.StatusOK, data)
}

func (htl *H5pController) GetContentUserDataByContentIdAndUser(c *gin.Context) {

	state := htl.svc.GetContentUserDataByContentIdAndUser(c)

	c.JSON(http.StatusOK, gin.H{
		"state": state,
	})
}

