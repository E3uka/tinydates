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
