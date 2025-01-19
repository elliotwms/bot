package bot

import (
	"net/http"
	"time"

	"github.com/elliotwms/bot/log"
)

func (bot *Bot) httpListen() {
	http.HandleFunc("/v1/health", func(w http.ResponseWriter, req *http.Request) {
		bot.log.Debug("Health check")
		latency := bot.session.HeartbeatLatency()
		// fail if we have not received a heartbeat response in more than 5 minutes
		if latency > 5*time.Minute {
			w.WriteHeader(500)
		}

		if _, err := w.Write([]byte(latency.String())); err != nil {
			bot.log.Error("Could not write health check response", log.WithErr(err))
		}
	})

	bot.log.Info("Serving health check", "endpoint", *bot.healthCheckAddr)

	err := http.ListenAndServe(*bot.healthCheckAddr, nil)
	if err != nil {
		bot.log.Error("Could not serve health check endpoint", log.WithErr(err))
		return
	}
}
