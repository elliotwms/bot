package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/interactions"
	"github.com/elliotwms/fakediscord/pkg/fakediscord"
	"github.com/neilotoole/slogt"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type RunStage struct {
	t       *testing.T
	require *require.Assertions

	session *discordgo.Session
	builder *Builder
	bot     *Bot

	ctx            context.Context
	cancel         context.CancelFunc
	runResult      error
	ready          *discordgo.Ready
	connected      bool
	healthAddr     string
	applicationID  string
	createdCommand *discordgo.ApplicationCommand
	guild          *discordgo.Guild
	channel        *discordgo.Channel
	handlerCalls   []*discordgo.InteractionCreate
	fakediscord    *fakediscord.Client
	interaction    *discordgo.InteractionCreate
}

func NewBotStage(t *testing.T) (*RunStage, *RunStage, *RunStage) {
	s := &RunStage{
		t:             t,
		require:       require.New(t),
		applicationID: appID,
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	session, err := discordgo.New("Bot token")
	s.require.NoError(err)
	s.session = session

	if os.Getenv("DEBUG_TEST") != "" {
		s.session.LogLevel = discordgo.LogDebug
		s.session.Debug = true
	}

	session.AddHandler(func(_ *discordgo.Session, e *discordgo.Connect) {
		s.connected = true
	})

	session.AddHandler(func(_ *discordgo.Session, e *discordgo.Disconnect) {
		s.connected = false
	})

	s.fakediscord = fakediscord.NewClient("test")

	// build the bot
	s.builder = New(s.applicationID, s.session).WithLogger(slogt.New(s.t))

	// create a default test guild
	s.guild, err = s.session.GuildCreate(s.t.Name())
	s.require.NoError(err)

	// create a default test channel
	s.channel, err = s.session.GuildChannelCreate(s.guild.ID, "test", discordgo.ChannelTypeGuildText)
	s.require.NoError(err)

	return s, s, s
}

func (s *RunStage) and() *RunStage {
	return s
}

func (s *RunStage) the_bot_has_custom_intents() {
	s.builder.WithIntents(discordgo.IntentGuilds)
}

func (s *RunStage) the_bot_has_a_health_check_endpoint() {
	port, err := freeport.GetFreePort()
	s.require.NoError(err)

	s.healthAddr = fmt.Sprintf("127.0.0.1:%d", port)

	s.builder.WithHealthCheck(s.healthAddr)
}

func (s *RunStage) a_ready_event_is_expected() *RunStage {
	cleanup := s.session.AddHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		s.ready = r
	})

	s.t.Cleanup(cleanup)

	return s
}

func (s *RunStage) the_bot_is_run() *RunStage {
	go func() {
		s.bot = s.builder.Build()

		s.runResult = s.bot.Run(s.ctx)
	}()

	return s
}

func (s *RunStage) the_ready_event_is_received() *RunStage {
	s.require.Eventually(func() bool {
		return s.ready != nil
	}, 1*time.Second, 10*time.Millisecond)

	return s
}

func (s *RunStage) the_ready_event_is_not_received() *RunStage {
	s.require.Never(func() bool {
		return s.ready != nil
	}, 1*time.Second, 10*time.Millisecond)

	return s
}

func (s *RunStage) the_bot_is_stopped() *RunStage {
	s.cancel()

	return s
}

func (s *RunStage) run_returns_no_error() *RunStage {
	s.require.NoError(s.runResult)

	return s
}

func (s *RunStage) an_established_session() *RunStage {
	s.require.NoError(s.session.Open())

	s.t.Cleanup(func() {
		s.require.NoError(s.session.Close())
	})

	return s
}

func (s *RunStage) the_session_is_connected() *RunStage {
	s.require.Eventually(func() bool {
		return s.connected
	}, time.Second, 10*time.Millisecond)

	return s
}

func (s *RunStage) the_session_is_disconnected() *RunStage {
	s.require.Eventually(func() bool {
		// discordgo sleeps for 1 second after disconnecting
		return !s.connected
	}, 5*time.Second, 100*time.Millisecond)

	return s
}

func (s *RunStage) the_session_has_custom_intents() {
	s.require.Equal(discordgo.IntentGuilds, s.session.Identify.Intents)
}

func (s *RunStage) the_health_check_should_succeed() {
	s.require.Eventually(func() bool {
		res, err := http.Get("http://" + s.healthAddr + "/v1/health")
		if err != nil {
			s.t.Log(err)
			return false
		}

		if res.StatusCode != 200 {
			s.t.Logf("Unexpected status code: %d", res.StatusCode)
			return false
		}

		return true
	}, 5*time.Second, 100*time.Millisecond)
}

// cleanup for tests unrelated to connectivity
func (s *RunStage) cleanup() {
	s.
		the_bot_is_stopped().and().
		the_session_is_disconnected()
}

func (s *RunStage) an_application_command_already_exists_named(name string) *RunStage {
	var err error
	s.createdCommand, err = s.session.ApplicationCommandCreate(s.applicationID, "", &discordgo.ApplicationCommand{
		Name: name,
		Type: discordgo.ChatApplicationCommand,
	})
	s.require.NoError(err)

	return s
}

func (s *RunStage) the_bot_has_application_command_named(name string) *RunStage {
	s.builder.WithApplicationCommand(&discordgo.ApplicationCommand{Name: name, Type: discordgo.ChatApplicationCommand}, func(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (err error) {
		s.t.Log("Handler called")
		s.handlerCalls = append(s.handlerCalls, i)

		return nil
	})

	return s
}

func (s *RunStage) a_command_should_exist_named(name string) *RunStage {
	s.require.Eventually(func() bool {
		commands, err := s.session.ApplicationCommands(s.applicationID, "")
		if err != nil {
			return false
		}

		for _, c := range commands {
			if c.Name == name {
				return true
			}
		}

		return false
	}, 1*time.Second, 100*time.Millisecond)

	return s
}

func (s *RunStage) a_command_should_not_exist_named(name string) *RunStage {
	s.require.Eventually(func() bool {
		commands, err := s.session.ApplicationCommands(s.applicationID, "")
		if err != nil {
			return false
		}

		for _, c := range commands {
			if c.Name == name {
				return false
			}
		}

		return true
	}, 1*time.Second, 100*time.Millisecond)

	return s
}

func (s *RunStage) migration_is_enabled() {
	s.builder.WithMigrationEnabled(true)
}

func (s *RunStage) the_command_is_invoked(name string) *RunStage {
	var err error
	s.interaction, err = s.fakediscord.Interaction(&discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		AppID: s.applicationID,
		Type:  discordgo.InteractionApplicationCommand,
		Data: discordgo.ApplicationCommandInteractionData{
			ID:          "id",
			Name:        name,
			CommandType: discordgo.ChatApplicationCommand,
		},
		ChannelID: s.channel.ID,
		GuildID:   s.guild.ID,
	}})

	s.require.NoError(err)

	return s
}

func (s *RunStage) the_bot_has_deferred_response_enabled() *RunStage {
	r := interactions.NewRouter(interactions.WithDeferredResponse(true))
	s.builder.WithRouter(r)

	return s
}

func (s *RunStage) the_interaction_should_have_a_deferred_response() {
	res, err := s.session.InteractionResponse(s.interaction.Interaction)
	s.require.NoError(err)

	s.require.Equal(discordgo.MessageTypeReply, res.Type)
	s.require.Equal(discordgo.MessageFlagsLoading, res.Flags)
}

func (s *RunStage) the_command_handler_should_have_been_called() *RunStage {
	s.require.Eventually(
		func() bool {
			return len(s.handlerCalls) > 0
		},
		1*time.Second,
		10*time.Millisecond,
		"the handler should have been called at least once",
	)

	return s
}
