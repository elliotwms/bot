package bot

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

type RunStage struct {
	t       *testing.T
	require *require.Assertions

	session    *discordgo.Session
	bot        *Bot
	ctx        context.Context
	cancel     context.CancelFunc
	runResult  error
	ready      *discordgo.Ready
	connected  bool
	healthAddr string
}

func NewRunStage(t *testing.T) (*RunStage, *RunStage, *RunStage) {
	s := &RunStage{
		t:       t,
		require: require.New(t),
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

	return s, s, s
}

func (s *RunStage) and() *RunStage {
	return s
}

func (s *RunStage) a_new_bot() *RunStage {
	s.bot = New("id", s.session).WithLogger(slog.Default())

	return s
}

func (s *RunStage) with_custom_intents() {
	s.bot.WithIntents(discordgo.IntentGuilds)
}

func (s *RunStage) with_health_check() {
	port, err := freeport.GetFreePort()
	s.require.NoError(err)

	s.healthAddr = fmt.Sprintf("127.0.0.1:%d", port)

	s.bot.WithHealthCheck(s.healthAddr)
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
