package bot

import (
	"log/slog"
	"maps"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/interactions/migrator"
	"github.com/elliotwms/bot/interactions/router"
	"github.com/elliotwms/bot/log"
)

// Builder builds a Bot.
// To enable application command migration, either set a specific migrator with WithMigrator or set
// WithMigrationEnabled and a default Migrator will be created based on the Bot's configuration.
type Builder struct {
	session          *discordgo.Session
	log              *slog.Logger
	applicationID    string
	healthCheckAddr  *string
	intents          discordgo.Intent
	router           *router.Router
	migrator         *migrator.Migrator
	guildID          string
	handlers         []interface{}
	commands         map[*discordgo.ApplicationCommand]router.ApplicationCommandHandler
	migrationEnabled bool
}

func New(applicationID string, session *discordgo.Session) *Builder {
	bot := &Builder{
		session:       session,
		log:           slog.New(log.DiscardHandler),
		applicationID: applicationID,
		commands:      make(map[*discordgo.ApplicationCommand]router.ApplicationCommandHandler),
	}

	return bot
}

func (b *Builder) WithLogger(l *slog.Logger) *Builder {
	b.log = l

	return b

}
func (b *Builder) WithIntents(i discordgo.Intent) *Builder {
	b.intents = i

	return b

}

func (b *Builder) WithHealthCheck(addr string) *Builder {
	b.healthCheckAddr = &addr

	return b
}

func (b *Builder) WithHandler(h interface{}) *Builder {
	b.handlers = append(b.handlers, h)

	return b
}

func (b *Builder) WithHandlers(hs []interface{}) *Builder {
	for _, h := range hs {
		b.WithHandler(h)
	}

	return b
}

func (b *Builder) WithRouter(r *router.Router) *Builder {
	b.router = r

	return b
}

func (b *Builder) WithMigrator(m *migrator.Migrator) *Builder {
	b.migrator = m

	return b
}

func (b *Builder) WithGuildID(guildID string) *Builder {
	b.guildID = guildID

	return b
}

func (b *Builder) WithMigrationEnabled(enabled bool) *Builder {
	b.migrationEnabled = enabled

	return b
}

func (b *Builder) WithApplicationCommand(c *discordgo.ApplicationCommand, h router.ApplicationCommandHandler) *Builder {
	b.commands[c] = h

	return b
}

func (b *Builder) WithApplicationCommands(handlers map[*discordgo.ApplicationCommand]router.ApplicationCommandHandler) *Builder {
	maps.Copy(b.commands, handlers)

	return b
}

func (b *Builder) Build() *Bot {
	bot := &Bot{
		session:         b.session,
		log:             b.log,
		applicationID:   b.applicationID,
		healthCheckAddr: b.healthCheckAddr,
		intents:         b.intents,
		router:          b.router,
		migrator:        b.migrator,
	}

	for _, h := range b.handlers {
		bot.handlerRemovers = append(bot.handlerRemovers, bot.session.AddHandler(h))
	}

	// register application commands with the router and migrator
	if len(b.commands) > 0 {
		if bot.router == nil {
			bot.router = router.New(router.WithLogger(bot.log))
		}

		if bot.migrator == nil && b.migrationEnabled {
			bot.migrator = migrator.New(
				b.session,
				b.applicationID,
				migrator.WithGuildID(b.guildID),
			)
		}

		for c, h := range b.commands {
			bot.router.RegisterCommand(c.Name, c.Type, h)

			if bot.migrator != nil {
				bot.migrator.WithApplicationCommand(c)
			}
		}
	}

	return bot
}
