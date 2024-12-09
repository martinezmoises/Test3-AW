package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// Health Check
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthCheckHandler)

	// Book Handlers
	router.HandlerFunc(http.MethodGet, "/api/v1/books", a.requireActivatedUser(a.listBooksHandler))          // List all books
	router.HandlerFunc(http.MethodGet, "/api/v1/books-search", a.requireActivatedUser(a.searchBooksHandler)) // Search books (new distinct route)
	router.HandlerFunc(http.MethodPost, "/api/v1/books", a.requireActivatedUser(a.createBookHandler))        // Add new book
	router.HandlerFunc(http.MethodGet, "/api/v1/books/:id", a.requireActivatedUser(a.displayBookHandler))    // Get book details
	router.HandlerFunc(http.MethodPut, "/api/v1/books/:id", a.requireActivatedUser(a.updateBookHandler))     // Update book details
	router.HandlerFunc(http.MethodDelete, "/api/v1/books/:id", a.requireActivatedUser(a.deleteBookHandler))  // Delete book

	// User Handlers
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)                                             // Register new user
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)                                    // Activate user
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)                // Authenticate token
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id", a.requireActivatedUser(a.getUserProfileHandler))            // Get user profile
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id/lists", a.requireActivatedUser(a.getUserReadingListsHandler)) // Get user's reading lists
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:id/reviews", a.requireActivatedUser(a.getUserReviewsHandler))    // Get user's reviews

	// Reading_lists handlers
	router.HandlerFunc(http.MethodGet, "/api/v1/lists", a.requireActivatedUser(a.listReadingListsHandler))
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:id", a.requireActivatedUser(a.getReadingListHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", a.requireActivatedUser(a.createReadingListHandler))
	router.HandlerFunc(http.MethodPut, "/api/v1/lists/:id", a.requireActivatedUser(a.updateReadingListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id", a.requireActivatedUser(a.deleteReadingListHandler))
	router.HandlerFunc(http.MethodPost, "/api/v1/lists/:id/books", a.requireActivatedUser(a.addBookToListHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:id/books", a.requireActivatedUser(a.removeBookFromListHandler))

	//Reviews handlers

	router.HandlerFunc(http.MethodGet, "/api/v1/books/:id/reviews", a.requireActivatedUser(a.listReviewsHandler))   // List reviews
	router.HandlerFunc(http.MethodPost, "/api/v1/books/:id/reviews", a.requireActivatedUser(a.createReviewHandler)) // Add review
	router.HandlerFunc(http.MethodPut, "/api/v1/reviews/:id", a.requireActivatedUser(a.updateReviewHandler))        // Update review
	router.HandlerFunc(http.MethodDelete, "/api/v1/reviews/:id", a.requireActivatedUser(a.deleteReviewHandler))     // Delete review

	// Password Reset Endpoints
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", a.createPasswordResetTokenHandler) // Generate password reset token
	router.HandlerFunc(http.MethodPut, "/v1/users/password", a.updateUserPasswordHandler)               // Reset password

	//Updated with Enabling CORS
	return a.recoverPanic(a.enableCORS(a.rateLimit(a.authenticate(router))))

}
