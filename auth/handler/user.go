package handler

import (
	"mediahub/auth/service"
	"mediahub/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, dto.ApiResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	auth, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Login successful",
		Data:    auth,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid user ID",
		})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User retrieved",
		Data:    user,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid user ID",
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Call service
	user, err := h.service.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User updated",
		Data:    user,
	})
}

// DeleteUser - DELETE /auth/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   "invalid user ID",
		})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User deleted",
	})
}
