package bot

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session         *discordgo.Session
	log             *slog.Logger
	applicationID   string
	healthCheckAddr *string
	intents         discordgo.Intent
	handlerRemovers []func()
}

func New(applicationID string, session *discordgo.Session) *Bot {
	bot := &Bot{
		session:       session,
		log:           slog.New(dh),
		applicationID: applicationID,
	}

	return bot
}

func (bot *Bot) WithSession(s *discordgo.Session) *Bot {
	bot.session = s

	return bot
}

func (bot *Bot) WithLogger(l *slog.Logger) *Bot {
	bot.log = l

	return bot
}

func (bot *Bot) WithIntents(i discordgo.Intent) *Bot {
	bot.intents = i

	return bot
}

func (bot *Bot) WithHandler(h interface{}) *Bot {
	bot.handlerRemovers = append(bot.handlerRemovers, bot.session.AddHandler(h))

	return bot
}

func (bot *Bot) WithHandlers(hs []interface{}) *Bot {
	for _, h := range hs {
		bot.WithHandler(h)
	}

	return bot
}

// Run runs the bot, starts the session if not already started, serves the health endpoint if present, and blocks
// until notified.
// If the session is already started, Run will not stop the session
func (bot *Bot) Run(ctx context.Context) error {
	if bot.intents != 0 {
		bot.session.Identify.Intents = bot.intents
	}

	bot.log.Info("Starting...")

	shouldCloseSession := true
	if err := bot.session.Open(); err != nil {
		if !errors.Is(err, discordgo.ErrWSAlreadyOpen) {
			bot.log.Error("Could not open session", withErr(err))

			return err
		}

		// session does not belong to bot, do not close it
		shouldCloseSession = false
	}

	if bot.healthCheckAddr != nil {
		go bot.httpListen()
	}

	<-ctx.Done()

	bot.log.Info("Stopping bot...")
	if shouldCloseSession {
		bot.log.Debug("Closing session")
		if err := bot.session.Close(); err != nil {
			bot.log.Error("Could not close session", withErr(err))
			return err
		}
	}

	for _, remove := range bot.handlerRemovers {
		remove()
	}

	return nil
}
