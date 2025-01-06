package bot

import (
	"testing"

	"github.com/elliotwms/fakediscord/pkg/fakediscord"
)

func TestMain(m *testing.M) {
	fakediscord.Configure("http://localhost:8080/")
	m.Run()
}

func TestNew(t *testing.T) {
	given, when, then := NewRunStage(t)

	given.
		a_new_bot().and().
		the_bot_watches_for_ready_events()

	when.
		the_bot_is_run()

	then.
		the_ready_event_is_received().and().
		the_session_is_connected()

	when.
		the_bot_is_stopped()

	then.
		run_returns_no_error().and().
		the_session_is_disconnected()
}

func TestNew_ShouldNotCloseEstablishedSession(t *testing.T) {
	given, when, then := NewRunStage(t)

	given.
		a_new_bot().and().
		an_established_session().and().
		the_session_is_connected().and().
		the_bot_watches_for_ready_events()

	when.
		the_bot_is_run()

	then.
		the_ready_event_is_not_received().
		the_bot_is_stopped().
		run_returns_no_error().and().
		the_session_is_connected()
}

func TestBot_WithIntents(t *testing.T) {
	given, when, then := NewRunStage(t)

	t.Cleanup(func() {
		then.
			the_bot_is_stopped().and().
			the_session_is_disconnected()
	})

	given.
		a_new_bot().
		with_custom_intents()

	when.
		the_bot_is_run()

	then.
		the_session_is_connected().and().
		the_session_has_custom_intents()
}

func TestBot_WithHealthCheck(t *testing.T) {
	given, when, then := NewRunStage(t)

	t.Cleanup(func() {
		then.
			the_bot_is_stopped().and().
			the_session_is_disconnected()
	})

	given.
		a_new_bot().
		with_health_check()

	when.
		the_bot_is_run()

	then.
		the_health_check_should_succeed()
}
