package lambda

import "testing"

func TestSessionFromParamStore(t *testing.T) {
	given, when, then := NewSessionStage(t)

	given.
		a_parameter_named_x_with_value_y("foo", "bar")

	when.
		a_new_session_from_param_store_is_requested_with_param_named("foo")

	then.
		no_error_should_be_returned().and().
		the_session_has_token("Bot bar")
}

func TestSessionFromParamStore_EmptyParamName(t *testing.T) {
	given, when, then := NewSessionStage(t)

	given.
		a_parameter_named_x_with_value_y("foo", "bar")

	when.
		a_new_session_from_param_store_is_requested_with_param_named("")

	then.
		an_error_should_be_returned("empty discord token paramstore parameter name")
}

func TestSessionFromParamStore_HttpError(t *testing.T) {
	given, when, then := NewSessionStage(t)

	given.
		the_param_store_server_is_unavailable()

	when.
		a_new_session_from_param_store_is_requested_with_param_named("foo")

	then.
		an_error_should_be_returned("failed to get parameter - http request error")
}

func TestSessionFromParamStore_EmptyParamValue(t *testing.T) {
	given, when, then := NewSessionStage(t)

	given.
		a_parameter_named_x_with_value_y("foo", "")

	when.
		a_new_session_from_param_store_is_requested_with_param_named("foo")

	then.
		an_error_should_be_returned("parameter empty")
}
