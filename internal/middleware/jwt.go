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

// JWTMiddleware 返回一个 Gin 中间件：
// - 验证 Authorization: Bearer <token>
// - 将解析后的用户名放到上下文（c.Set("user" , username)）
func JWTMiddleware(cfg config.JWTConfig) gin.HandlerFunc {
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
		tok, err := p.Parse(tokStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(cfg.Secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims, ok := tok.Claims.(jwt.MapClaims); ok {
			if sub, ok := claims["sub"].(string); ok {
				c.Set("user", sub)
			}
		}
		c.Next()
	}
}
