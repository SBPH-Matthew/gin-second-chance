package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			// Ensure we still allow CORS headers to be sent even on failure
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication cookie missing"})
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "dev-secret-key"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// IMPORTANT: Verify the signing method to prevent "none" algorithm exploits
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		// token.Valid already checks the 'exp' claim automatically
		// if it exists in the payload. No need for manual time check.
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID in token"})
			return
		}

		c.Set("user_id", uint(userIDFloat))
		c.Next()
	}
}

// func AuthRequired() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// 1. Get the token from the cookie instead of the header
// 		tokenString, err := c.Cookie("token")
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication cookie missing"})
// 			c.Abort()
// 			return
// 		}

// 		secret := os.Getenv("JWT_SECRET")
// 		if secret == "" {
// 			secret = "dev-secret-key"
// 		}

// 		// 2. Parse the token (This part remains largely the same)
// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
// 			return []byte(secret), nil
// 		})

// 		if err != nil || !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
// 			c.Abort()
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
// 			c.Abort()
// 			return
// 		}

// 		// Note: jwt.Parse with Valid check already handles 'exp' internally,
// 		// but keeping your manual check is fine for extra safety.
// 		if exp, ok := claims["exp"].(float64); ok {
// 			if time.Now().Unix() > int64(exp) {
// 				c.JSON(http.StatusUnauthorized, gin.H{"message": "Token has expired"})
// 				c.Abort()
// 				return
// 			}
// 		}

// 		userIDFloat, ok := claims["user_id"].(float64)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID in token"})
// 			c.Abort()
// 			return
// 		}

// 		c.Set("user_id", uint(userIDFloat))
// 		c.Next()
// 	}
// }
