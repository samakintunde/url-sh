package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/auth"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/user"
	"url-shortener/internal/utils"
	"url-shortener/internal/validation"
)

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
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
					"errors": utils.ErrorResponse(errors.New(http.StatusText(http.StatusBadRequest))),
				})
				return
			}

			if errs := validator.Validate(req); errs != nil {
				slog.Error(handlerID, "message", "couldn't validate signup request", "errors", errs)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			createUserArgs := user.CreateUserParams{
				Email:     req.Email,
				Password:  req.Password,
				FirstName: req.FirstName,
				LastName:  req.LastName,
			}

			createdUser, err := userService.CreateUser(ctx, createUserArgs)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't create user", "error", err)

				switch err {
				case user.ErrPasswordWeak:
					fallthrough
				case user.ErrPasswordCompromised:
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

			slog.Info(handlerID, "message", "created user", "user", createdUser.ID)

			createEmailVerificationArgs := emailverification.CreateEmailVerificationParams{
				UserID:    createdUser.ID,
				UserEmail: createdUser.Email,
			}

			err = emailVerificationService.CreateEmailVerification(ctx, createEmailVerificationArgs)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't create email verification", "error", err)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			slog.Info(handlerID, "message", "created email verification", "email", createdUser.Email)

			data := newUserResponse(createdUser)

			utils.RespondWithJSON(w, http.StatusCreated, map[string]any{
				"data": data,
			})
		},
	)
}

func HandleVerifyEmail(ctx context.Context, validator validation.Validator, userService user.UserService, emailVerificationService emailverification.EmailVerificationService) http.Handler {
	handlerID := "handler.user.HandleVerifyEmail"

	type request struct {
		Email string `json:"email" validate:"required,email"`
		Code  string `json:"code" validate:"required,len=8,alphanum"`
	}

	type response struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
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
				slog.Error(handlerID, "message", "couldn't validate signup request", "errors", errs)
				utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			getUserArgs := user.GetUserByEmailParams{
				Email: req.Email,
			}

			userData, err := userService.GetUserByEmail(ctx, getUserArgs)

			if err != nil {
				if err == user.ErrUserNotFound {
					slog.Info(handlerID, "message", "user not found", "email", req.Email, "error", err)
					utils.RespondWithJSON(w, http.StatusNotFound, map[string]any{
						"errors": []string{"no pending verification found for that account"},
					})
					return
				}
				slog.Error(handlerID, "error getting existing user", "error", err)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			hasUserCompletedEmailVerificationArgs := emailverification.HasUserCompletedEmailVerificationParams{
				UserID:    userData.ID,
				UserEmail: userData.Email,
			}

			hasCompletedVerification, err := emailVerificationService.HasUserCompletedEmailVerification(ctx, hasUserCompletedEmailVerificationArgs)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't confirm email verification status", "error", err, "email", userData.Email)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"couldn't confirm email verification status"},
				})
				return
			}

			if hasCompletedVerification {
				slog.Info(handlerID, "message", "email already verified")
				utils.RespondWithJSON(w, http.StatusAlreadyReported, map[string]any{
					"errors": []string{"email already verified"},
				})
				return
			}

			completeEmailVerificationArgs := emailverification.CompleteEmailVerificationParams{
				UserID:    userData.ID,
				UserEmail: req.Email,
				Code:      req.Code,
			}

			err = emailVerificationService.CompleteEmailVerification(ctx, completeEmailVerificationArgs)

			if err != nil {
				if err == emailverification.ErrInvalidVerificationCode {
					slog.Error(handlerID, "message", "code is invalid", "error", err)
					utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
						"errors": []string{"code is invalid"},
					})
					return
				}
				slog.Error(handlerID, "message", "couldn't verify email", "error", err, "email", userData.Email)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"couldn't verify email"},
				})
				return
			}

			data := response{
				ID:            userData.ID,
				Email:         userData.Email,
				EmailVerified: true,
			}

			utils.RespondWithJSON(w, http.StatusOK, map[string]any{
				"data": data,
			})
		})
}

func HandleLogin(ctx context.Context, validator validation.Validator, tokenMaker token.Maker, userService *user.UserService, emailVerificationService *emailverification.EmailVerificationService, emailer email.Emailer) http.Handler {
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

			getUserArgs := user.GetUserByEmailParams{
				Email: req.Email,
			}

			userData, err := userService.GetUserByEmail(ctx, getUserArgs)

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

			doPasswordsMatch, err := auth.VerifyPassword(req.Password, userData.Password)

			if err != nil {
				slog.Error(handlerID, "message", "Error verifying password", "error", err)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			if !doPasswordsMatch {
				slog.Error(handlerID, "message", "incorrect password", "email", req.Email)
				utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{
					"errors": []string{"Incorrect username or password"},
				})
				return
			}

			hasUserCompletedEmailVerificationArgs := emailverification.HasUserCompletedEmailVerificationParams{
				UserID:    userData.ID,
				UserEmail: userData.Email,
			}

			hasCompletedVerification, err := emailVerificationService.HasUserCompletedEmailVerification(ctx, hasUserCompletedEmailVerificationArgs)

			fmt.Println(hasCompletedVerification, err)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't confirm email verification status", "error", err, "email", userData.Email)
				utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"failed to confirm email verification status"},
				})
				return
			}

			if !hasCompletedVerification {
				slog.Info(handlerID, "message", "email not verified")
				args := emailverification.CreateEmailVerificationParams{
					UserID:    userData.ID,
					UserEmail: userData.Email,
				}
				emailVerification, err := emailVerificationService.RecreateEmailVerification(ctx, args)
				if err != nil {
					slog.Error(handlerID, "message", "couldn't recreate email verification", "error", err, "email", userData.Email)
					utils.RespondWithJSON(w, http.StatusNotModified, map[string]any{
						"errors": []string{http.StatusText(http.StatusInternalServerError)},
					})
					return
				}
				emailer.Send([]string{userData.Email}, "Verify your account", fmt.Sprintf("Your email verification code is %s", emailVerification.Code))
				utils.RespondWithJSON(w, http.StatusForbidden, map[string]any{
					"errors": []string{"You must verify your email to continue"},
				})
				return
			}

			token, _, err := tokenMaker.CreateToken(userData.ID, 24*time.Hour)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't create token", "email", req.Email)
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
				User:        newUserResponse(userData),
			}

			updateLoginTimeArgs := user.UpdateLoginTimeParams{
				ID:   userData.ID,
				Time: time.Now(),
			}

			err = userService.UpdateLastLogin(ctx, updateLoginTimeArgs)

			if err != nil {
				slog.Error(handlerID, "message", "couldn't update login time", "email", userData.Email, "time", updateLoginTimeArgs.Time)
			}

			// Prevents tokens from referrer leakage
			w.Header().Set("Referrer-Policy", "strict-origin")
			utils.RespondWithJSON(w, http.StatusOK, map[string]any{
				"data": data,
			})
		},
	)
}

func HandleStartResetPassword(ctx context.Context, validator validation.Validator, userService *user.UserService, emailer email.Emailer) http.Handler {
	handlerID := "handler.user.HandleStartResetPassword"
	type request struct {
		Email string `json:"email" validate:"required,email"`
	}
	type response struct {
		ExpiresAt time.Time `json:"expires_at"`
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

		startResetPasswordArgs := user.StartPasswordResetParams{
			Email: req.Email,
		}

		passwordResetToken, err := userService.StartPasswordReset(ctx, startResetPasswordArgs)

		if err != nil {
			slog.Error(handlerID, "message", "couldn't create start password reset", "error", err)
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{
				"errors": []string{err.Error()},
			})
			return
		}

		err = emailer.Send([]string{req.Email}, "Reset your password", fmt.Sprintf("Your reset token is %s", passwordResetToken.Token))

		if err != nil {
			slog.Error(handlerID, "message", "couldn't send password reset email", "error", err)
		}

		data := response{
			ExpiresAt: passwordResetToken.ExpiresAt,
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{
			"data": data,
		})
	})
}

func HandleResetPassword(ctx context.Context, validator validation.Validator, userService *user.UserService, emailer email.Emailer) http.Handler {
	handlerID := "handler.user.HandleStartResetPassword"
	type request struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=8,max=50,password"`
	}
	type response struct {
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

		passwordResetToken, err := userService.ResetPassword(ctx, resetPasswordArgs)

		if err != nil {
			if err == user.ErrReusingPassword {
				slog.Error(handlerID, "error", err, "user", passwordResetToken.UserID)
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

		getUserArgs := user.GetUserByIDParams{
			ID: passwordResetToken.UserID,
		}

		user, err := userService.GetUserByID(ctx, getUserArgs)

		if err != nil {
		} else {
			emailer.Send([]string{user.Email}, "Password reset", "Your password has been reset. Contact support if it wasn't you.")
		}

		data := response{}

		// Prevents tokens from referrer leakage
		w.Header().Set("Referrer-Policy", "strict-origin")
		utils.RespondWithJSON(w, http.StatusOK, map[string]any{
			"data": data,
		})
	})
}
