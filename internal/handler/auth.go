package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"net/http"
)

func (h *Handler) signUp(c *gin.Context) {
	input := new(domain.User)

	if err := c.BindJSON(input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *Handler) signIn(c *gin.Context) {

}
