package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"time"
	db "url-shortener/db/sqlc"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
)

func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "All systems OK")
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

func HandleSignup(ctx context.Context, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator, email Emailer) http.Handler {
	type request struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=8,max=50,password"`
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	type response struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	validateSignUpRequest := func(req request) []ValidationError {
		err := validate.Struct(req)
		if err != nil {
			var errors []ValidationError

			for _, err := range err.(validator.ValidationErrors) {
				field, _ := reflect.TypeOf(req).FieldByName(err.Field())
				jsonFieldName := getJSONFieldName(field)
				var message string
				switch err.Tag() {
				case "required":
					message = fmt.Sprintf("%s is required", jsonFieldName)
				case "email":
					message = err.Translate(trans)
				case "min":
					message = fmt.Sprintf("%s must be at least %s characters long", jsonFieldName, err.Param())
				case "max":
					message = fmt.Sprintf("%s must be at most %s characters long", jsonFieldName, err.Param())
				case "password":
					message = err.Translate(trans)
				default:
					message = fmt.Sprintf("%s is invalid", jsonFieldName)
				}

				errors = append(errors, ValidationError{
					Field:   jsonFieldName,
					Message: message,
				})
			}
			return errors
		}
		return nil
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Decode request body
			req, err := decode[request](r)

			if err != nil {
				slog.Error("Invalid Request Payload", "error", err)
				respondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": []string{http.StatusText(http.StatusBadRequest)},
				})
				return
			}

			// Validate request
			if errs := validateSignUpRequest(req); errs != nil {
				slog.Error("couldn't validate signup request", "errors", errs)
				respondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			// Check for existing account
			exists, err := queries.DoesUserExistByEmail(ctx, req.Email)
			if err != nil {
				slog.Error("database error querying existing user", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			if exists == 1 {
				slog.Error("account already exists", "email", req.Email, "error", err)
				respondWithJSON(w, http.StatusConflict, map[string]any{
					"errors": []string{"Account already exists"},
				})
				return
			}

			// Check Password Strength
			isPasswordStrong := checkPasswordStrength(req.Password)

			if !isPasswordStrong {
				slog.Info("weak password", "email", req.Email)
				respondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": []string{"Password is too weak. Please use a stronger password."},
				})
				return
			}

			// Check if Password is being re-used or has been leaked
			// Losing about 250ms on this call to the haveibeenpwned API so disabling for now
			//
			// isPasswordPwned, err := checkPasswordPwned(req.Password)

			// if err != nil {
			// 	slog.Info("weak password", "email", req.Email)
			// 	respondWithJSON(w, http.StatusInternalServerError, map[string]any{
			// 		"errors": []string{"Something went wrong. Try again."},
			// 	})
			// 	return
			// }

			// if isPasswordPwned {
			// 	slog.Info("reusing leaked password", "email", req.Email)
			// 	respondWithJSON(w, http.StatusBadRequest, map[string]any{
			// 		"errors": []string{"Password has been exposed in data breaches. Please use a different password."},
			// 	})
			// 	return
			// }

			// Generate ULID
			id := NewULID()

			// Hash password
			hashedPassword, err := HashPassword(req.Password)

			if err != nil {
				slog.Error("Error generating hash", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			userParams := db.CreateUserParams{
				ID:       id.String(),
				Email:    req.Email,
				Password: hashedPassword,
				FirstName: sql.NullString{
					String: req.FirstName,
					Valid:  true,
				},
				LastName: sql.NullString{
					String: req.LastName,
					Valid:  true,
				},
			}

			// TODO: Wrap database actions in a transaction

			// Create User in DB
			createdUser, err := queries.CreateUser(ctx, userParams)

			if err != nil {
				if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					slog.Error("couldn't create user with email", "email", userParams.Email, "error", err)
					respondWithJSON(w, http.StatusConflict, map[string]any{
						"errors": []string{"Account already exists"},
					})
					return
				}
				slog.Error("database error creating user", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}
			slog.Info("created user", "user", createdUser.ID)

			// Send verification e-mail
			code, err := GenerateVerificationToken()

			if err != nil {
				slog.Error("couldn't generate verification token", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			emailVerificationParam := db.CreateEmailVerificationParams{
				UserID:    createdUser.ID,
				Email:     createdUser.Email,
				Code:      code,
				ExpiresAt: time.Now().Add(time.Minute * 15),
			}
			err = queries.CreateEmailVerification(ctx, emailVerificationParam)

			if err != nil {
				slog.Error("couldn't create email verification", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}
			slog.Info("created email verification", "email", createdUser.Email, "code", code)

			err = email.Send([]string{createdUser.Email}, "Verify your Account", fmt.Sprintf("Your verification code is: %s", code))

			if err != nil {
				slog.Error("couldn't send email verification", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			// Serve success response
			data := response{
				ID:        createdUser.ID,
				Email:     createdUser.Email,
				FirstName: createdUser.FirstName.String,
				LastName:  createdUser.LastName.String,
			}
			respondWithJSON(w, http.StatusCreated, map[string]any{
				"data": data,
			})
		},
	)
}

func HandleVerifyEmail(ctx context.Context, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator) http.Handler {
	type request struct {
		Email string `json:"email" validate:"required,email"`
		Code  string `json:"code" validate:"required,len=8,alphanum"`
	}

	type response struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	validateVerifyEmailRequest := func(req request) []ValidationError {
		err := validate.Struct(req)
		if err != nil {
			var errors []ValidationError

			for _, err := range err.(validator.ValidationErrors) {
				field, _ := reflect.TypeOf(req).FieldByName(err.Field())
				jsonFieldName := getJSONFieldName(field)
				var message string
				switch err.Tag() {
				case "required":
					message = fmt.Sprintf("%s is required", jsonFieldName)
				case "email":
					message = err.Translate(trans)
				case "len":
					message = fmt.Sprintf("%s must be %s characters long", jsonFieldName, err.Param())
				case "alphanum":
					message = fmt.Sprintf("%s must be alphanumeric characters only", jsonFieldName)
				default:
					message = fmt.Sprintf("%s is invalid", jsonFieldName)
				}

				errors = append(errors, ValidationError{
					Field:   jsonFieldName,
					Message: message,
				})
			}
			return errors
		}
		return nil
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Decode request body
			req, err := decode[request](r)

			if err != nil {
				slog.Error("Invalid Request Payload", "error", err)
				respondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": []string{http.StatusText(http.StatusBadRequest)},
				})
				return
			}

			// Validate request
			if errs := validateVerifyEmailRequest(req); errs != nil {
				slog.Error("couldn't validate signup request", "errors", errs)
				respondWithJSON(w, http.StatusBadRequest, map[string]any{
					"errors": errs,
				})
				return
			}

			user, err := queries.GetUserByEmail(ctx, req.Email)
			if err != nil {
				if err == sql.ErrNoRows {
					slog.Error("user not found", "email", req.Email, "error", err)
					respondWithJSON(w, http.StatusNotFound, map[string]any{
						"errors": []string{http.StatusText(http.StatusNotFound)},
					})
					return
				}
				slog.Error("database error querying existing user", "error", err)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{http.StatusText(http.StatusInternalServerError)},
				})
				return
			}

			if user.ID == "" {
				slog.Error("account not found", "email", req.Email, "error", err)
				respondWithJSON(w, http.StatusConflict, map[string]any{
					"errors": []string{"Account not found"},
				})
				return
			}

			hasCompletedVerification, err := queries.IsUserEmailVerificationComplete(ctx, db.IsUserEmailVerificationCompleteParams{
				UserID: user.ID,
				Email:  user.Email,
			})

			if err != nil {
				slog.Error("couldn't confirm email verification status", "error", err, "email", user.Email)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"couldn't confirm email verification status"},
				})
				return
			}

			if hasCompletedVerification == 1 {
				slog.Error("email already verified")
				respondWithJSON(w, http.StatusAlreadyReported, map[string]any{
					"errors": []string{"email already verified"},
				})
				return
			}

			emailVerification, err := queries.CompleteEmailVerification(ctx, db.CompleteEmailVerificationParams{
				UserID: user.ID,
				Email:  req.Email,
				Code:   req.Code,
			})

			if err != nil {
				if err == sql.ErrNoRows {
					slog.Error("code is invalid", "error", err, "email", user.Email, "code", req.Code)
					respondWithJSON(w, http.StatusBadRequest, map[string]any{
						"errors": []string{"code is invalid"},
					})
					return
				}
				slog.Error("couldn't verify email", "error", err, "email", user.Email)
				respondWithJSON(w, http.StatusInternalServerError, map[string]any{
					"errors": []string{"couldn't verify email"},
				})
				return
			}

			data := response{
				ID:            user.ID,
				Email:         user.Email,
				EmailVerified: emailVerification.VerifiedAt.Valid,
			}

			respondWithJSON(w, http.StatusOK, map[string]any{
				"data": data,
			})
		})
}

func HandleLogin(ctx context.Context, queries *db.Queries) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Login")
		},
	)
}
