package api

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/service"
)

type UserServicer interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	IsEmailAvailable(ctx context.Context, email string) error
	Register(ctx context.Context, req dto.RegisterReqeust) (*dto.LoginResponse, error)
	SendEmailCode(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.ForgetPasswordReqeust) (*dto.LoginResponse, error)
}

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	userService UserServicer
}

func NewUserHandler(userService UserServicer) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数无效，请检查邮箱和密码格式",
		})
		return
	}

	resp, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNameOrPasswordFailed) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误，请稍后重试",
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterReqeust
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数无效，请检查邮箱、密码和验证码",
		})
		return
	}

	if err := h.userService.IsEmailAvailable(c.Request.Context(), req.Email); err != nil {
		if errors.Is(err, service.ErrEmailAleadyExist) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "该邮箱已注册",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误，请稍后重试",
		})
		return
	}

	resp, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailCodeNotEqual) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "邮箱验证码错误",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误，请稍后重试",
		})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) ForgetPassword(c *gin.Context) {
	var req dto.ForgetPasswordReqeust
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数无效，请检查邮箱、密码和验证码",
		})
		return
	}

	// 邮箱必须已存在才能重置密码
	if err := h.userService.IsEmailAvailable(c.Request.Context(), req.Email); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "该邮箱未注册",
		})
		return
	} else if !errors.Is(err, service.ErrEmailAleadyExist) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误，请稍后重试",
		})
		return
	}

	resp, err := h.userService.ResetPassword(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailCodeNotEqual) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "邮箱验证码错误",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误，请稍后重试",
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) SendEmailCode(c *gin.Context) {
	email := c.Param("email")
	log.Printf("email: %s", email)

	if err := h.userService.SendEmailCode(c.Request.Context(), email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "验证码发送失败，请稍后重试",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

var _ UserServicer = (*service.UserService)(nil)
