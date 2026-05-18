package middleware

import (
	"be-Clever School/i18n"

	"github.com/gin-gonic/gin"
)

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.GetHeader("X-Locale")
		if locale == "" {
			locale = "vi"
		}
		i18n.SetCurrentLocale(locale)
		c.Next()
	}
}
