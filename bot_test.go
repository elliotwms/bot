package bot

import "testing"

func TestNew(t *testing.T) {
	given, when, then := NewBotStage(t)

	given.
		a_ready_event_is_expected()

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
	given, when, then := NewBotStage(t)

	given.
		an_established_session().and().
		the_session_is_connected().and().
		a_ready_event_is_expected()

	when.
		the_bot_is_run()

	then.
		the_ready_event_is_not_received().
		the_bot_is_stopped().
		run_returns_no_error().and().
		the_session_is_connected()
}

func TestBot_WithIntents(t *testing.T) {
	given, when, then := NewBotStage(t)

	t.Cleanup(then.cleanup)

	given.
		the_bot_has_custom_intents()

	when.
		the_bot_is_run()

	then.
		the_session_is_connected().and().
		the_session_has_custom_intents()
}

func TestBot_WithHealthCheck(t *testing.T) {
	given, when, then := NewBotStage(t)

	t.Cleanup(then.cleanup)

	given.
		the_bot_has_a_health_check_endpoint()

	when.
		the_bot_is_run()

	then.
		the_health_check_should_succeed()
}

func TestBot_WithCommand(t *testing.T) {
	given, when, then := NewBotStage(t)

	t.Cleanup(then.cleanup)

	given.
		an_application_command_already_exists_named("foo").and().
		the_bot_has_application_command_named("bar").and().
		migration_is_enabled()

	when.
		the_bot_is_run()

	then.
		a_command_should_not_exist_named("foo").and().
		a_command_should_exist_named("bar")
}

func TestBot_WithCommand_MigrationDisabled(t *testing.T) {
	given, when, then := NewBotStage(t)

	t.Cleanup(then.cleanup)

	given.
		an_application_command_already_exists_named("baz").and().
		the_bot_has_application_command_named("bux")

	when.
		the_bot_is_run()

	then.
		a_command_should_exist_named("baz").and().
		a_command_should_not_exist_named("bux")
}

func TestBot_WithCommand_WithDeferredResponse(t *testing.T) {
	given, when, then := NewBotStage(t)

	t.Cleanup(then.cleanup)

	given.
		the_bot_has_deferred_response_enabled().and().
		the_bot_has_application_command_named("foo").and().
		migration_is_enabled()

	when.
		the_bot_is_run().and().
		the_session_is_connected().
		the_command_is_invoked("foo")

	then.
		the_command_handler_should_have_been_called().and().
		the_interaction_should_have_a_deferred_response()
}
