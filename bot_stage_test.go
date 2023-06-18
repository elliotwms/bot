package bot

import (
	"context"
	"github.com/sirupsen/logrus/hooks/test"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type RunStage struct {
	t       *testing.T
	require *require.Assertions

	session   *discordgo.Session
	bot       *Bot
	ctx       context.Context
	cancel    context.CancelFunc
	runResult error
	log       *logrus.Logger
	ready     *discordgo.Ready
	connected bool
	loghook   *test.Hook
}

func NewRunStage(t *testing.T) (*RunStage, *RunStage, *RunStage) {
	s := &RunStage{
		t:       t,
		require: require.New(t),
		log:     logrus.New(),
	}

	s.loghook = test.NewLocal(s.log)

	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.log.SetOutput(os.Stdout)

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

	return s, s, s
}

func (s *RunStage) and() *RunStage {
	return s
}

func (s *RunStage) a_new_bot() *RunStage {
	s.bot = New("id", s.session, s.log)

	return s
}

func (s *RunStage) with_custom_intents() {
	s.bot = s.bot.WithIntents(discordgo.IntentGuilds)
}

func (s *RunStage) the_bot_watches_for_ready_events() *RunStage {
	s.bot.WithHandler(func(_ *discordgo.Session, r *discordgo.Ready) {
		s.ready = r
	})

	return s
}

func (s *RunStage) the_bot_is_run() *RunStage {
	go func() {
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

func (s *RunStage) with_config_reporter_with_custom_value(k, v string) {
	s.bot = s.bot.WithConfigReporter(func(showSensitive bool) logrus.Fields {
		return logrus.Fields{
			k: v,
		}
	})
}

func (s *RunStage) the_starting_message_is_logged_with_field(k, v string) {
	found := false
	for _, entry := range s.loghook.AllEntries() {
		if entry.Message == "Starting..." {
			if entry.Data[k] == v {
				found = true
			}
		}
	}

	s.require.True(found)
}
