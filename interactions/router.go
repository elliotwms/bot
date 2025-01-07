package interactions

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/log"
)

type ApplicationCommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (err error)

type Router struct {
	applicationCommandHandlers map[string]ApplicationCommandHandler
	log                        *slog.Logger
	deferredResponseEnabled    bool
}

func NewRouter(options ...func(*Router)) *Router {
	r := &Router{
		applicationCommandHandlers: make(map[string]ApplicationCommandHandler),
		log:                        slog.New(log.DiscardHandler),
	}

	for _, o := range options {
		o(r)
	}

	return r
}

func WithLogger(l *slog.Logger) func(*Router) {
	return func(r *Router) {
		r.log = l
	}
}

func WithDeferredResponse(enabled bool) func(*Router) {
	return func(r *Router) {
		r.deferredResponseEnabled = enabled
	}
}

func WithCommandHandler(name string, handler ApplicationCommandHandler) func(*Router) {
	return func(r *Router) {
		r.RegisterCommand(name, handler)
	}
}

func (r *Router) RegisterCommand(name string, handler ApplicationCommandHandler) {
	r.applicationCommandHandlers[name] = handler
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

	h, ok := r.applicationCommandHandlers[command.Name]
	if !ok {
		r.log.Error("Handler not found for application command", "name", command.Name)
		return
	}

	if err := h(s, e, command); err != nil {
		r.log.Error("Failed to handle interaction", "error", err)
	}
}
