package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"net/http"
	"time"
)

// ------------------ Create Account ------------------------
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type createUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var errPq *pq.Error
		if errors.As(err, &errPq) {
			switch errPq.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusCreated, response)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if err = util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(user.Username,
		server.config.AccessTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshTokenPayload, err := server.tokenMaker.CreateToken(user.Username,
		server.config.RefreshTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           uuid.New(),
		Username:     req.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		ExpiresAt:    refreshTokenPayload.ExpiredAt,
	})

	if err != nil {
		var errPq *pq.Error
		if errors.As(err, &errPq) {
			switch errPq.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("the refresh token is blocked")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		err := fmt.Errorf("the refresh token has expired")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	responseToken := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, responseToken)
}
