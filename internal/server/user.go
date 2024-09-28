package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"url-shortener/internal/auth"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/user"
	"url-shortener/internal/utils"
	"url-shortener/internal/validation"
)

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

func newUserResponse(user user.User) userResponse {
	return userResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

func HandleSignup(ctx context.Context, validator validation.Validator, userService user.UserService, emailVerificationService emailverification.EmailVerificationService) http.Handler {
	handlerID := "handler.user.HandleSignup"

	type request struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=8,max=50,password"`
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	type response userResponse

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := utils.DecodeToJSON[request](r)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't decode request payload", "error", err)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"error": utils.ErrorResponse(errors.New(http.StatusText(http.StatusBadRequest))),
				})
				return
			}

			if errs := validator.Validate(req); errs != nil {
				slog.Error(handlerID, "message", "couldn't validate request", "errors", errs)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			createUserArgs := user.RegisterUserParams{
				Email:     req.Email,
				Password:  req.Password,
				FirstName: req.FirstName,
				LastName:  req.LastName,
			}

			createdUser, err := userService.RegisterUser(ctx, createUserArgs)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't create user", "error", err)

				switch err {
				case auth.ErrWeakPassword:
					fallthrough
				case auth.ErrCompromisedPassword:
					fallthrough
				case user.ErrUserExists:
					utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
						"errors": []string{err.Error()},
					})
					return
				default:
					utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
						"errors": []string{http.StatusText(http.StatusInternalServerError)},
					})
					return
				}

			}

			utils.RespondWithJSON(w, http.StatusCreated, map[string]any{
				"data": newUserResponse(createdUser),
			})
		},
	)
}

func HandleVerifyEmail(ctx context.Context, validator validation.Validator, userService user.UserService, emailVerificationService emailverification.EmailVerificationService) http.Handler {
	handlerID := "handler.user.HandleVerifyEmail"

	type request struct {
		Code string `json:"code" validate:"required"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := utils.DecodeToJSON[request](r)

			if err != nil {
				slog.Error(handlerID, "message", "Invalid Request Payload", "error", err)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"error": http.StatusText(http.StatusBadRequest),
				})
				return
			}

			if errs := validator.Validate(req); errs != nil {
				slog.Error(handlerID, "message", "couldn't validate request", "errors", errs)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			verifyEmailArgs := user.VerifyEmailParams{
				Code: req.Code,
			}

			err = userService.VerifyEmail(ctx, verifyEmailArgs)

			if err != nil {
				if err == emailverification.ErrInvalidVerificationCode {
					slog.Error(handlerID, "message", "code is invalid", "error", err)
					utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
						"errors": []string{"code is invalid"},
					})
					return
				}
				slog.Error(handlerID, "message", "couldn't verify email", "error", err)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"couldn't verify email"},
				})
				return
			}

			utils.RespondWithJSON(w, http.StatusOK, map[string]any{})
		})
}

func HandleLogin(ctx context.Context, validator validation.Validator, tokenMaker token.Maker, userService *user.UserService, emailVerificationService *emailverification.EmailVerificationService) http.Handler {
	handlerID := "handler.user.HandleLogin"

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	type response struct {
		AccessToken string       `json:"access_token"`
		User        userResponse `json:"user"`
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := utils.DecodeToJSON[request](r)

			if err != nil {
				slog.Error(handlerID, "message", "Invalid Request Payload", "error", err)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": []string{http.StatusText(http.StatusBadRequest)},
				})
				return
			}

			if errs := validator.Validate(req); errs != nil {
				slog.Error("couldn't validate login request", "errors", errs)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			loginUserArgs := user.LoginUserParams{
				Email:    req.Email,
				Password: req.Password,
			}

			loggedInUser, err := userService.LoginUser(ctx, loginUserArgs)

			if err != nil {
				if err == user.ErrUserNotFound {
					slog.Error(handlerID, "message", "user not found", "email", req.Email, "error", err)
					// TODO: fix user enumeration attack
					utils.RespondWithJSON(w, http.StatusNotFound, map[string]any{
						"errors": []string{"Incorrect username or password"},
					})
					return
				}
				slog.Error(handlerID, "message", "database error querying existing user", "error", err)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			token, _, err := tokenMaker.CreateToken(loggedInUser.ID, 24*time.Hour)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't create token")
				utils.RespondWithJSON(w, http.StatusNotImplemented, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			cookie := &http.Cookie{
				Name:     "access_token",
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(time.Hour * 24),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(w, cookie)

			data := response{
				AccessToken: token,
				User:        newUserResponse(loggedInUser),
			}

			// Prevents tokens from referrer leakage
			w.Header().Set("Referrer-Policy", "strict-origin")
			utils.RespondWithJSON(w, http.StatusOK, map[string]any{
				"data": data,
			})
		},
	)
}

func HandleStartResetPassword(ctx context.Context, validator validation.Validator, userService *user.UserService) http.Handler {
	handlerID := "handler.user.HandleStartResetPassword"

	type request struct {
		Email string `json:"email" validate:"required,email"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := utils.DecodeToJSON[request](r)

		if err != nil {
			slog.Error(handlerID, "message", "Invalid Request Payload", "error", err)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": []string{http.StatusText(http.StatusBadRequest)},
			})
			return
		}

		if errs := validator.Validate(req); errs != nil {
			slog.Error(handlerID, "message", "couldn't validate login request", "errors", errs)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": errs,
			})
			return
		}

		startPasswordResetArgs := user.StartPasswordResetParams{
			Email: req.Email,
		}

		err = userService.StartPasswordReset(ctx, startPasswordResetArgs)

		if err != nil {
			slog.Error(handlerID, "message", "couldn't create start password reset", "error", err)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": []string{err.Error()},
			})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{})
	})
}

func HandleResetPassword(ctx context.Context, validator validation.Validator, userService *user.UserService) http.Handler {
	handlerID := "handler.user.HandleStartResetPassword"
	type request struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=8,max=50,password"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := utils.DecodeToJSON[request](r)

		if err != nil {
			slog.Error(handlerID, "message", "Invalid Request Payload", "error", err)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": []string{http.StatusText(http.StatusBadRequest)},
			})
			return
		}

		if errs := validator.Validate(req); errs != nil {
			slog.Error(handlerID, "message", "couldn't validate login request", "errors", errs)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": errs,
			})
			return
		}

		resetPasswordArgs := user.ResetPasswordParams{
			Token:    req.Token,
			Password: req.Password,
		}

		err = userService.ResetPassword(ctx, resetPasswordArgs)

		if err != nil {
			if err == user.ErrReusingPassword {
				slog.Error(handlerID, "error", err)
				utils.RespondWithJSON(w, http.StatusForbidden, map[string]any{
					"errors": []string{"can't reuse current password"},
				})
				return
			}
			slog.Error(handlerID, "message", "couldn't reset password", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
				"errors": []string{http.StatusText(http.StatusInternalServerError)},
			})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{})
	})
}
