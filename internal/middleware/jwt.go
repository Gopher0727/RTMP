package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/Gopher0727/RTMP/config"
)

// GenerateToken 使用配置生成基于用户名的 JWT（HMAC SHA256）。
func GenerateToken(username string, cfg config.JWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"iss": cfg.Issuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(cfg.AccessExpMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// JWTAuth JWT认证中间件
// @Summary JWT认证中间件
// @Description 验证请求头中的JWT令牌
// @Tags middleware
// @Accept json
// @Produce json
// @Security BearerAuth
// @Failure 401 {object} map[string]string
// @Router /api/v1/* [get]
func JWTAuth() gin.HandlerFunc {
	cfg := config.GetJWTConfig()
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.Fields(auth)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		tokStr := parts[1]
		p := &jwt.Parser{}
		tok, err := p.Parse(tokStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		username, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
