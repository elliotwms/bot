package interactions

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestMigrator_NewMigrator(t *testing.T) {
	s, _ := discordgo.New("token")

	m := NewMigrator(s, "foo")

	require.Empty(t, m.guildID)
}

func TestMigrator_NewMigrator_WithGuildID(t *testing.T) {
	s, _ := discordgo.New("token")

	m := NewMigrator(s, "foo", WithGuildID("bar"))

	require.Equal(t, "bar", m.guildID)
}

func TestMigrator_Migrate(t *testing.T) {
	tests := map[string]string{
		"no guild id":   "",
		"with guild id": "guildid",
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			updater := &testCommandUpdater{}
			m := &Migrator{
				s:       updater,
				appID:   "foo",
				guildID: v,
			}

			m.WithApplicationCommand(&discordgo.ApplicationCommand{})

			err := m.Migrate(context.Background())

			require.NoError(t, err)
			require.Equal(t, 1, updater.calls)
			require.Equal(t, "foo", updater.appID)
			require.Equal(t, v, updater.guildID)
			require.Len(t, updater.commands, 1)
		})
	}
}

type testCommandUpdater struct {
	calls    int
	appID    string
	guildID  string
	commands []*discordgo.ApplicationCommand
}

func (t *testCommandUpdater) ApplicationCommandBulkOverwrite(appID, guildID string, commands []*discordgo.ApplicationCommand, opts ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	t.calls++

	t.appID, t.guildID = appID, guildID
	t.commands = commands

	return commands, nil
}
