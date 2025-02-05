package migrator

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type commandUpdater interface {
	ApplicationCommandBulkOverwrite(appID, guildID string, commands []*discordgo.ApplicationCommand, opts ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
}

// Migrator migrates application commands for a given application. It can also be provided with a guild ID (see
// WithGuildID), which can be used during testing to create the commands within a test guild instead.
type Migrator struct {
	s        commandUpdater
	appID    string
	commands []*discordgo.ApplicationCommand
	guildID  string
}

type Option func(*Migrator)

// New creates a new Migrator.
func New(s *discordgo.Session, appID string, opts ...Option) *Migrator {
	m := &Migrator{s: s, appID: appID}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithGuildID configures a guild to create the commands in (as opposed to creating them globally).
func WithGuildID(id string) Option {
	return func(migrator *Migrator) {
		migrator.guildID = id
	}
}

// WithApplicationCommand registers an application command to be migrated when Migrate is called.
func (m *Migrator) WithApplicationCommand(c *discordgo.ApplicationCommand) *Migrator {
	m.commands = append(m.commands, c)

	return m
}

// Migrate migrates the application's commands.
func (m *Migrator) Migrate(ctx context.Context) error {
	_, err := m.s.ApplicationCommandBulkOverwrite(m.appID, m.guildID, m.commands, discordgo.WithContext(ctx))

	return err
}
