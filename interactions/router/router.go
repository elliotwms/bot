package router

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/log"
)

type ApplicationCommandHandler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (err error)

type key struct {
	name        string
	commandType discordgo.ApplicationCommandType
}

type Router struct {
	applicationCommandHandlers map[key]ApplicationCommandHandler
	log                        *slog.Logger
	deferredResponseEnabled    bool
}

type Option func(*Router)

func New(options ...func(*Router)) *Router {
	r := &Router{
		applicationCommandHandlers: make(map[key]ApplicationCommandHandler),
		log:                        slog.New(log.DiscardHandler),
	}

	for _, o := range options {
		o(r)
	}

	return r
}

func WithLogger(l *slog.Logger) Option {
	return func(r *Router) {
		r.log = l
	}
}

// WithDeferredResponse adds an initial deferred response to command invocations
func WithDeferredResponse(enabled bool) Option {
	return func(r *Router) {
		r.deferredResponseEnabled = enabled
	}
}

func (r *Router) RegisterCommand(name string, commandType discordgo.ApplicationCommandType, handler ApplicationCommandHandler) {
	r.applicationCommandHandlers[key{name: name, commandType: commandType}] = handler
}

// Handle implements the discordgo.InteractionCreate handler, dispatching events to the relevant handlers within the
// router. Currently only application commands are supported
func (r *Router) Handle(s *discordgo.Session, e *discordgo.InteractionCreate) {
	_ = r.HandleWithContext(context.Background(), s, e)
}

// HandleWithContext propagates the context and provides a request/response pattern for interaction handling (e.g. via Lambda)
func (r *Router) HandleWithContext(ctx context.Context, is *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.InteractionResponse {
	// todo support other interaction types i.e. message component, autocomplete, modal submit
	switch i.Type {
	case discordgo.InteractionPing:
		return &discordgo.InteractionResponse{Type: discordgo.InteractionResponsePong}
	case discordgo.InteractionApplicationCommand:
		r.handleApplicationCommand(ctx, is, i)
		return nil
	default:
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "Unexpected interaction"},
		}
	}
}

func (r *Router) handleApplicationCommand(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) {
	// if deferred response is enabled then call discord with the initial response before routing the command
	if r.deferredResponseEnabled {
		err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		}, discordgo.WithContext(ctx))
		if err != nil {
			r.log.Error("Failed to respond to InteractionCreate", "error", err)
			return
		}
	}

	command := e.ApplicationCommandData()

	h, ok := r.applicationCommandHandlers[key{command.Name, command.CommandType}]
	if !ok {
		r.log.Error("Handler not found for application command", "name", command.Name)
		return
	}

	if err := h(ctx, s, e, command); err != nil {
		r.log.Error("Failed to handle interaction", "error", err)
	}
}
