package bot

import (
	"net/http"
	"time"
)

func (bot *Bot) WithHealthCheck(addr string) *Bot {
	bot.healthCheckAddr = &addr

	return bot
}

func (bot *Bot) httpListen() {
	http.HandleFunc("/v1/health", func(w http.ResponseWriter, req *http.Request) {
		bot.log.Debug("Health check")
		latency := bot.session.HeartbeatLatency()
		// fail if we have not received a heartbeat response in more than 5 minutes
		if latency > 5*time.Minute {
			w.WriteHeader(500)
		}

		if _, err := w.Write([]byte(latency.String())); err != nil {
			bot.log.Error("Could not write health check response", withErr(err))
		}
	})

	bot.log.Info("Serving health check", "endpoint", *bot.healthCheckAddr)

	err := http.ListenAndServe(*bot.healthCheckAddr, nil)
	if err != nil {
		bot.log.Error("Could not serve health check endpoint", withErr(err))
		return
	}
}
