package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/llaoj/kube-finder/internal/config"
	log "github.com/sirupsen/logrus"
)

var signingKey = []byte("2i#L7Hym@2#O1")

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// tokenString := c.Request.Header.Get("Authorization")
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "token not found"})
			c.Abort()
			return
		}
		log.WithFields(log.Fields{"authorization": tokenString}).Trace()
		// tokenString = strings.Fields(tokenString)[1]
		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("userID", token.Claims.(*jwt.StandardClaims).Audience)
		log.WithFields(log.Fields{"context.keys": c.Keys}).Trace()

		c.Next()
	}
}

func NewJWT(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 60*30, // 30min
		Issuer:    config.Get().Name,
		Audience:  userID,
	})
	ss, _ := token.SignedString(signingKey)
	return ss
}
