package interactions

import (
	"log/slog"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

type RouterStage struct {
	t       testing.TB
	require *require.Assertions

	router        *Router
	handlerCalled int
}

func NewRouterStage(t *testing.T) (*RouterStage, *RouterStage, *RouterStage) {
	s := &RouterStage{
		t:       t,
		require: require.New(t),
		router:  NewRouter(WithLogger(slog.Default())),
	}

	return s, s, s
}

func (s *RouterStage) a_handler_is_registered_for_command(name string) {
	s.router.RegisterCommand(name, discordgo.ChatApplicationCommand, func(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (err error) {
		s.handlerCalled++

		return nil
	})
}

func (s *RouterStage) the_router_is_called_for_command(name string) {
	s.router.Handle(nil, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name:        name,
				CommandType: discordgo.ChatApplicationCommand,
			},
		},
	})
}

func (s *RouterStage) the_handler_should_have_been_called_n_times(i int) {
	s.require.Equal(i, s.handlerCalled)
}
