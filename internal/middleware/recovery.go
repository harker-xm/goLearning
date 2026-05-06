package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-programming-tour-book/blog-service/global"
	"github.com/go-programming-tour-book/blog-service/pkg/app"
	"github.com/go-programming-tour-book/blog-service/pkg/email"
	"github.com/go-programming-tour-book/blog-service/pkg/errcode"
)

// Recovery recovers from panics, logs the error, sends an email alert, and returns a 500 response.
func Recovery() gin.HandlerFunc {
	defFB := func(c *gin.Context, err interface{}) {
		if global.EmailSetting != nil {
			mailer := email.NewEmail(&email.SMTPInfo{
				Host:     global.EmailSetting.Host,
				Port:     global.EmailSetting.Port,
				IsSSL:    global.EmailSetting.IsSSL,
				UserName: global.EmailSetting.UserName,
				Password: global.EmailSetting.Password,
				From:     global.EmailSetting.From,
			})
			_ = mailer.SendMail(
				global.EmailSetting.To,
				fmt.Sprintf("异常抛出，发生时间: %d", time.Now().Unix()),
				fmt.Sprintf("错误信息: %v", err),
			)
		}
		app.NewResponse(c).ToErrorResponse(errcode.ServerError)
		c.Abort()
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				global.Logger.WithCallersFrames().Errorf("panic recover err: %v", err)
				defFB(c, err)
			}
		}()
		c.Next()
	}
}
