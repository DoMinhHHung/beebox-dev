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

func RegisterAuthRoutes(rg *gin.RouterGroup, svc *auth.Service) {
	rg.POST("/sign-up", signUpHandler(svc))
	rg.POST("/sign-in", signInHandler(svc))
	rg.POST("/sign-out", signOutHandler(svc))
}

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
