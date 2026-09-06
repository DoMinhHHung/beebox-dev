package httpapi

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	"github.com/gin-gonic/gin"
)

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInResponse struct {
	Token string `json:"token"`
}

// RegisterAuthRoutes đăng ký các endpoint HTTP cho đăng ký, đăng nhập và đăng xuất người dùng.
func RegisterAuthRoutes(rg *gin.RouterGroup, svc *auth.Service) {
	rg.POST("/sign-up", signUpHandler(svc))
	rg.POST("/sign-in", signInHandler(svc))
	rg.POST("/sign-out", signOutHandler(svc))
}

// signUpHandler tạo HTTP handler xử lý yêu cầu đăng ký tài khoản và trả về thông tin người dùng đã tạo. 
// Handler trả về trạng thái 201 cùng ID và email của người dùng khi đăng ký thành công.
func signUpHandler(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req signUpRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "invalid request body"))
			return
		}

		o, err := svc.SignUp(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, signUpResponse{ID: o.ID, Email: o.Email})
	}
}

// signInHandler tạo HTTP handler xử lý yêu cầu đăng nhập và trả về token xác thực.
func signInHandler(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req signInRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(apperror.New(apperror.CodeInvalidInput, "invalid request body"))
			return
		}

		token, err := svc.SignIn(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, signInResponse{Token: token})
	}
}

// signOutHandler tạo HTTP handler để đăng xuất người dùng bằng bearer token trong header Authorization.
// Handler trả về 204 nếu thiếu hoặc không hợp lệ, chuyển lỗi đăng xuất hợp lệ vào context,
// và trả về 204 khi đăng xuất thành công.
func signOutHandler(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Status(http.StatusNoContent)
			return
		}

		token, err := parseBearerToken(header)
		if err != nil {
			c.Status(http.StatusNoContent)
			return
		}

		if err := svc.SignOut(c.Request.Context(), token); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
