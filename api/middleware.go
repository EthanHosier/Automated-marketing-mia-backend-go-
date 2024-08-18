package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethanhosier/mia-backend-go/utils"
	"github.com/golang-jwt/jwt/v4"
	// "github.com/nedpals/supabase-go"
)

type Middleware func(http.Handler) http.Handler

func CreateMiddlewareStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			next = xs[i](next)
		}
		return next
	}
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		log.Printf("Recieved %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(wrapped, r)
		log.Printf("Completed operation %v %s %s in %v", wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

// func Auth(supabaseClient *supabase.Client, next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		token := strings.TrimPrefix(authHeader, "Bearer ")

// 		user, err := supabaseClient.Auth.User(r.Context(), token)
// 		if err != nil {
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), utils.UserIdKey, user.ID)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtSecret = os.Getenv("SUPABASE_JWT_SECRET")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token method is HMAC and matches the secret
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract user ID from token claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println(claims)

		userID, ok := claims["sub"].(string)
		log.Println("id: ", userID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), utils.UserIdKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
