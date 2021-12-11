package transport

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"net/http"
	"time"
)

// LoggingMiddleware logs every http requests
func LoggingMiddleware(logger zerolog.Logger, next http.Handler) http.Handler {
	var h = hlog.NewHandler(logger)
	var accessHandler = hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Send()
	})

	var handler = h(
		accessHandler(
			hlog.RemoteAddrHandler("ip")(
				hlog.UserAgentHandler("user_agent")(
					hlog.RefererHandler("referer")(
						hlog.RequestIDHandler("req_id", "Request-Id")(
							next,
						),
					),
				),
			),
		),
	)

	return handler
}
