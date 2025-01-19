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
	migrator        *interactions.Migrator
}

// Run runs the bot, starts the session if not already started, serves the health endpoint if present, and blocks
// until context is done.
// If the session is already started, Run will not stop the session
func (bot *Bot) Run(ctx context.Context) error {
	if bot.intents != 0 {
		bot.session.Identify.Intents = bot.intents
	}

	// add the router handler for InteractionCreate events
	if bot.router != nil {
		bot.handlerRemovers = append(bot.handlerRemovers, bot.session.AddHandler(bot.router.Handle))
	}

	if bot.migrator != nil {
		bot.log.Info("Migrating application commands")
		err := bot.migrator.Migrate(ctx)
		if err != nil {
			bot.log.Error("Failed to migrate application commands: " + err.Error())
			return err
		}
	}

	bot.log.Info("Starting")

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

	bot.log.Info("Shutting down")
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
