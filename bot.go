package bot

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/interactions"
	"github.com/elliotwms/bot/log"
)

type Bot struct {
	session         *discordgo.Session
	log             *slog.Logger
	applicationID   string
	healthCheckAddr *string
	intents         discordgo.Intent
	handlerRemovers []func()
	router          *interactions.Router
}

func New(applicationID string, session *discordgo.Session, options ...func(*Bot)) *Bot {
	bot := &Bot{
		session:       session,
		log:           slog.New(log.DiscardHandler),
		applicationID: applicationID,
		router:        interactions.NewRouter(),
	}

	for _, o := range options {
		o(bot)
	}

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

func (bot *Bot) WithApplicationCommand(name string, h interactions.ApplicationCommandHandler) *Bot {
	bot.router.RegisterCommand(name, h)

	return bot
}

// Run runs the bot, starts the session if not already started, serves the health endpoint if present, and blocks
// until context is completed.
// If the session is already started, Run will not stop the session
func (bot *Bot) Run(ctx context.Context) error {
	if bot.intents != 0 {
		bot.session.Identify.Intents = bot.intents
	}

	// add the InteractionCreate handler
	// todo make conditional depending on if commands are registered
	bot.WithHandler(bot.router.Handle)

	bot.log.Info("Starting...")

	shouldCloseSession := true
	if err := bot.session.Open(); err != nil {
		if !errors.Is(err, discordgo.ErrWSAlreadyOpen) {
			bot.log.Error("Could not open session", log.WithErr(err))

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
			bot.log.Error("Could not close session", log.WithErr(err))
			return err
		}
	}

	for _, remove := range bot.handlerRemovers {
		remove()
	}

	return nil
}
