package interactions

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/log"
)

type ApplicationCommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (err error)

type key struct {
	name        string
	commandType discordgo.ApplicationCommandType
}

type Router struct {
	applicationCommandHandlers map[key]ApplicationCommandHandler
	log                        *slog.Logger
	deferredResponseEnabled    bool
}

type RouterOption func(*Router)

func NewRouter(options ...func(*Router)) *Router {
	r := &Router{
		applicationCommandHandlers: make(map[key]ApplicationCommandHandler),
		log:                        slog.New(log.DiscardHandler),
	}

	for _, o := range options {
		o(r)
	}

	return r
}

func WithLogger(l *slog.Logger) RouterOption {
	return func(r *Router) {
		r.log = l
	}
}

// WithDeferredResponse adds an initial deferred response to command invocations
func WithDeferredResponse(enabled bool) RouterOption {
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
	// todo handle other interaction types
	switch e.Type {
	case discordgo.InteractionApplicationCommand:
		r.handleApplicationCommand(s, e)
	default:
		r.log.Error("Unexpected interaction type", "type", e.Type.String())
	}
}

func (r *Router) handleApplicationCommand(s *discordgo.Session, e *discordgo.InteractionCreate) {
	// if deferred response is enabled then call discord with the initial response before routing the command
	if r.deferredResponseEnabled {
		err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
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

	if err := h(s, e, command); err != nil {
		r.log.Error("Failed to handle interaction", "error", err)
	}
}
