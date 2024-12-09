package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/martinezmoises/Test3/internal/data"
	"github.com/martinezmoises/Test3/internal/validator"
)

func (a *applicationDependencies) registerUserHandler(w http.ResponseWriter,
	r *http.Request) {
	// Get the passed in data from the request body and store in a temporary struct
	var incomingData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, &incomingData)

	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// we will add the password later after we have hashed it
	user := &data.User{
		Username:  incomingData.Username,
		Email:     incomingData.Email,
		Activated: false,
	}

	// hash the password and store it along with the cleartext version
	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	// Perform validation for the User
	v := validator.New()

	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.userModel.Insert(user) // we will add userModel to main later
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return

	}

	token, err := a.tokenModel.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := map[string]any{
		"activationToken": token.Plaintext,
		"userID":          user.ID,
	}

	err = a.mailer.Send(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		a.logger.Error(err.Error())
	}

	// Status code 201 resource created
	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the body from the request and store in temporary struct
	var incomingData struct {
		TokenPlaintext string `json:"token"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// Validate the data
	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.TokenPlaintext)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Let's check if the token provided belongs to the user
	// We will implement the GetForToken() method later
	user, err := a.userModel.GetForToken(data.ScopeActivation, incomingData.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true
	err = a.userModel.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflictResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// User has been activated so let's delete the activation token to
	// prevent reuse.
	err = a.tokenModel.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send a response
	data := envelope{
		"user": user,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

//new handlers

func (a *applicationDependencies) getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	user, err := a.userModel.GetByID(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getUserReadingListsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	lists, err := a.readingListModel.GetByUserID(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reading_lists": lists}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getUserReviewsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	reviews, err := a.reviewModel.GetByUserID(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"reviews": reviews}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// Create a password reset token
func (a *applicationDependencies) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Email string `json:"email"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, incomingData.Email)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := a.userModel.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// Send a generic response to prevent enumeration
			a.genericResponse(w, r, http.StatusOK, "an email will be sent to you containing password reset instructions")
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Ensure the user is activated
	if !user.Activated {
		a.failedValidationResponse(w, r, map[string]string{"email": "this account is not activated"})
		return
	}

	// Generate token
	token, err := a.tokenModel.New(user.ID, 1*time.Hour, data.ScopePasswordReset)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send reset email
	emailData := map[string]any{
		"passwordResetToken": token.Plaintext,
		"userID":             user.ID,
	}

	// Send reset email
	err = a.mailer.Send(user.Email, "password_reset.tmpl", emailData)
	if err != nil {
		// Corrected logging format for slog.Logger
		a.logger.Error("failed to send password reset email",
			slog.String("email", user.Email),
			slog.String("error", err.Error()),
		)
	}

	// Respond to the client
	a.genericResponse(w, r, http.StatusOK, "an email will be sent to you containing password reset instructions")
}

// Reset the user's password
func (a *applicationDependencies) updateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidatePasswordPlaintext(v, incomingData.Password)
	data.ValidateTokenPlaintext(v, incomingData.Token)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Verify token and fetch user
	user, err := a.userModel.GetForToken(data.ScopePasswordReset, incomingData.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.failedValidationResponse(w, r, map[string]string{"token": "invalid or expired token"})
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update user's password
	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.userModel.Update(user)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Delete all password reset tokens
	err = a.tokenModel.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	a.genericResponse(w, r, http.StatusOK, "your password was successfully reset")
}
