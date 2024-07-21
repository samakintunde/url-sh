package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
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

func HandleSignup(ctx context.Context, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator) http.Handler {
	type SignUpRequest struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=8"`
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	type SignUpResponse struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	validateSignUpRequest := func(req SignUpRequest) []ValidationError {
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
					message = "Invalid email format"
				case "min":
					message = fmt.Sprintf("%s must be at least %s characters long", jsonFieldName, err.Param())
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
			req, err := decode[SignUpRequest](r)

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

			user := db.CreateUserParams{
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

			// Create User in DB
			createdUser, err := queries.CreateUser(ctx, user)

			if err != nil {
				if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					slog.Error("couldn't create user with email", "email", user.Email, "error", err)
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

			// Serve success response
			// TODO: hide sensitive details like password from logs
			slog.Info("created user", "user", createdUser)
			respondWithJSON(w, http.StatusCreated, map[string]any{
				"data": SignUpResponse{
					ID:        createdUser.ID,
					Email:     createdUser.Email,
					FirstName: createdUser.FirstName.String,
					LastName:  createdUser.LastName.String,
				},
			})
		},
	)
}

func HandleLogin(ctx context.Context, queries *db.Queries) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Login")
		},
	)
}
