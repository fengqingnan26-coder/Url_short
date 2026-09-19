package middleware

// FIXME

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/jekyulll/url_shortener/pkg/logger"
// 	"go.uber.org/zap"
// )

// func Logger() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		raw := c.Request.URL.RawQuery

// 		c.Next()

// 		end := time.Now()
// 		latency := end.Sub(start)

// 		req := c.Request
// 		res := c.Writer

// 		fields := []zap.Field{
// 			zap.String("remote_ip", c.ClientIP()),
// 			zap.String("latency", latency.String()),
// 			zap.String("host", req.Host),
// 			zap.String("request", req.Method+" "+path),
// 			zap.Int("status", res.Status()),
// 			zap.Int64("size", int64(res.Size())),
// 			zap.String("user_agent", req.UserAgent()),
// 		}

// 		if raw != "" {
// 			path = path + "?" + raw
// 		}

// 		id := req.Header.Get("X-Request-ID")
// 		if id != "" {
// 			fields = append(fields, zap.String("request_id", id))
// 		}

// 		n := res.Status()
// 		switch {
// 		case n >= http.StatusInternalServerError:
// 			logger.Error("Server error", fields...)
// 		case n >= http.StatusBadRequest:
// 			logger.Warn("Client error", fields...)
// 		case n >= http.StatusMultipleChoices:
// 			logger.Info("Redirection", fields...)
// 		default:
// 			logger.Info("Success", fields...)
// 		}
// 	}
// }
