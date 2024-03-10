package tinydates

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewTinydatesHandler(ctx context.Context, svc Service) http.Handler {
	router := gin.Default()

	router.Use(contentTypeJSON())

	router.GET("/user/create", func(c *gin.Context) {
		user, err := svc.CreateUser(ctx)
		if err != nil {
			// an internal server errror must have occured
			c.JSON(
				http.StatusInternalServerError,
				GenericErrResponse{Err: err.Error()},
			)
		}

		c.JSON(http.StatusCreated, user)
	})

	router.POST("/login", func(c *gin.Context) {
		var request LoginRequest

		// deserialize JSON POST request into the LoginRequest struct, if
		// serialization fails return a `GenericErrResponse` to the caller with
		// the appropriate status code for bad request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, GenericErrResponse{Err: err.Error()})
			return
		}

		// submit a login request to the service, for simplicity any errors
		// found from this point is treated like an invalid login attempt
		loginResponse, err := svc.Login(ctx, request)
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				GenericErrResponse{Err: err.Error()},
			)
			return
		}

		c.JSON(http.StatusCreated, loginResponse)
	})

	return router
}

// contentTypeJSON middleware sets the response header to application/json for
// all subequent routes it has been applied to.
func contentTypeJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/json")
		ctx.Next()
	}
}
