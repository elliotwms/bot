package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"os"
)

type Bot struct {
	session         *discordgo.Session
	log             *logrus.Logger
	applicationID   string
	healthCheckAddr *string
	configReporter  func(showSensitive bool) logrus.Fields
	intents         discordgo.Intent
}

func New(applicationID string, session *discordgo.Session, l *logrus.Logger) *Bot {
	bot := &Bot{
		session:       session,
		log:           l,
		applicationID: applicationID,
	}

	return bot
}

func (bot *Bot) WithSession(s *discordgo.Session) *Bot {
	bot.session = s

	return bot
}

func (bot *Bot) WithIntents(i discordgo.Intent) *Bot {
	bot.intents = i

	return bot
}

func (bot *Bot) WithConfigReporter(f func(showSensitive bool) logrus.Fields) *Bot {
	bot.configReporter = f

	return bot
}

func (bot *Bot) WithHandler(h interface{}) *Bot {
	bot.session.AddHandler(h)

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
func (bot *Bot) Run(notify <-chan os.Signal) error {
	if bot.intents != 0 {
		bot.session.Identify.Intents = bot.intents
	}

	fields := logrus.Fields{}
	if bot.configReporter != nil {
		fields = bot.configReporter(bot.log.IsLevelEnabled(logrus.TraceLevel))
	}
	bot.log.WithFields(fields).Info("Starting...")

	shouldCloseSession := true
	if err := bot.session.Open(); err != nil {
		if !errors.Is(err, discordgo.ErrWSAlreadyOpen) {
			bot.log.
				WithError(err).
				Error("Could not open session")

			return err
		}

		// session does not belong to bot, do not close it
		shouldCloseSession = false
	}

	if bot.healthCheckAddr != nil {
		go bot.httpListen()
	}

	<-notify

	bot.log.Info("Stopping bot...")
	if shouldCloseSession {
		bot.log.Debug("Closing session")
		if err := bot.session.Close(); err != nil {
			bot.log.WithError(err).Error("Could not close session")
			return err
		}
	}

	return nil
}
