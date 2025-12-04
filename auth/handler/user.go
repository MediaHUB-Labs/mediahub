package handler

import (
	"mediahub/auth/service"
	"mediahub/dto"
	"mediahub/utils"
	"net/http"

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
		utils.LogToFile(err.Error())
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, dto.ApiResponse{
				Success: false,
				Error:   err.Error(),
			})
			utils.LogToFile(err.Error())
			return
		}
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}
	utils.LogToFile("User Creation Successful")

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
			Error:   utils.ParseValidationError(err),
		})
		utils.LogToFile(err.Error())
		return
	}

	auth, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}
	utils.LogToFile("Login Successful!")
	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "Login successful",
		Data:    auth,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	var req dto.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}

	id := req.ID
	user, err := h.service.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}

	utils.LogToFile("User Data Retrieved")

	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User retrieved",
		Data:    user,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}

	id := req.ID
	// Call service
	user, err := h.service.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}
	utils.LogToFile("User Data Updated")
	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User updated",
		Data:    user,
	})
}

// DeleteUser - DELETE /auth/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req dto.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}

	id := req.ID
	if err := h.service.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		utils.LogToFile(err.Error())
		return
	}
	utils.LogToFile("User Deleted")
	c.JSON(http.StatusOK, dto.ApiResponse{
		Success: true,
		Message: "User deleted",
	})
}
