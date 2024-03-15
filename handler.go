package tinydates

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func NewTinydatesHandler(ctx context.Context, svc Service) http.Handler {
	handler := gin.Default()

	handler.Use(contentTypeJSON())

	handler.GET("/user/create", func(c *gin.Context) {
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

	handler.POST("/login", func(c *gin.Context) {
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
		switch err {
		case nil:
		case ErrUnauthorized:
			c.JSON(
				http.StatusUnauthorized,
				GenericErrResponse{Err: err.Error()},
			)
			return
		default:
			c.JSON(
				http.StatusInternalServerError,
				GenericErrResponse{Err: err.Error()},
			)
			return
		}

		c.JSON(http.StatusCreated, loginResponse)
	})

	handler.GET("/discover", func(c *gin.Context) {
		idString := c.GetHeader("Id")
		token := c.GetHeader("Authorization")
		minAge, minAgeSupplied := c.GetQuery("minAge")
		maxAge, maxAgeSupplied := c.GetQuery("maxAge")
		byPopularity, byPopularitySupplied := c.GetQuery("orderByPopularity")

		id, err := strconv.Atoi(idString)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				GenericErrResponse{Err: err.Error()},
			)
			return
		}

		orderByPopularity := false
		if byPopularitySupplied {
			orderByPopularity, err = strconv.ParseBool(byPopularity)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					GenericErrResponse{Err: err.Error()},
				)
				return
			}

		}

		users, err := svc.Discover(
			ctx,
			id,
			token,
			minAge,
			minAgeSupplied,
			maxAge,
			maxAgeSupplied,
			orderByPopularity,
		)
		switch err {
		case nil:
		case ErrUnauthorized:
			c.JSON(
				http.StatusUnauthorized,
				GenericErrResponse{Err: err.Error()},
			)
			return
		case ErrUnauthorized:
			c.JSON(
				http.StatusUnauthorized,
				GenericErrResponse{Err: err.Error()},
			)
			return
		case ErrMinOrMaxAgeMissing:
			c.JSON(
				http.StatusBadRequest,
				GenericErrResponse{Err: err.Error()},
			)
			return
		case ErrMinOrMaxAgeInvalid:
			c.JSON(
				http.StatusBadRequest,
				GenericErrResponse{Err: err.Error()},
			)
			return
		case ErrorMinOrMaxFormat:
			c.JSON(
				http.StatusBadRequest,
				GenericErrResponse{Err: err.Error()},
			)
			return
		default:
			c.JSON(
				http.StatusInternalServerError,
				GenericErrResponse{Err: err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, users)
	})

	handler.POST("/swipe", func(c *gin.Context) {
		var request SwipeRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, GenericErrResponse{Err: err.Error()})
			return
		}

		token := c.GetHeader("Authorization")

		response, err := svc.Swipe(ctx, token, request)
		switch err {
		case nil:
		case ErrUnauthorized:
			c.JSON(
				http.StatusUnauthorized,
				GenericErrResponse{Err: err.Error()},
			)
			return
		default:
			c.JSON(
				http.StatusInternalServerError,
				GenericErrResponse{Err: err.Error()},
			)
			return
		}

		c.JSON(http.StatusOK, response)
	})

	return handler
}

// contentTypeJSON middleware sets the response header to application/json for
// all subequent routes it has been applied to.
func contentTypeJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/json")
		ctx.Next()
	}
}
