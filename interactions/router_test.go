package interactions

import (
	"testing"
)

func TestRouter_ApplicationCommand(t *testing.T) {
	given, when, then := NewRouterStage(t)

	given.
		a_handler_is_registered_for_command("foo")

	when.
		the_router_is_called_for_command("foo")

	then.
		the_handler_should_have_been_called_n_times(1)

}

func TestRouter_ApplicationCommand_NotFound(t *testing.T) {
	given, when, then := NewRouterStage(t)

	given.
		a_handler_is_registered_for_command("foo")

	when.
		the_router_is_called_for_command("bar")

	then.
		the_handler_should_have_been_called_n_times(0)
}
