package lambda

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot/interactions/router"
	"github.com/elliotwms/bot/log"
)

const (
	headerSignature = "X-Signature-Ed25519"
	headerTimestamp = "X-Signature-Timestamp"
)

type Endpoint struct {
	s         *discordgo.Session
	publicKey ed25519.PublicKey
	router    *router.Router
	log       *slog.Logger
}

func New(publicKey ed25519.PublicKey, options ...Option) *Endpoint {
	logger := slog.New(log.DiscardHandler)

	e := &Endpoint{
		publicKey: publicKey,
		log:       logger,
		router:    router.New(router.WithLogger(logger)),
	}

	for _, o := range options {
		o(e)
	}

	return e
}

type Option func(*Endpoint)

// WithRouter overrides the underlying router used for the endpoint.
func WithRouter(router *router.Router) Option {
	return func(endpoint *Endpoint) {
		endpoint.router = router
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(endpoint *Endpoint) {
		endpoint.log = logger
	}
}

// WithSession adds a global session. This is provided instead of the interaction-specific discordgo.Session.
func (r *Endpoint) WithSession(s *discordgo.Session) *Endpoint {
	r.s = s

	return r
}

// WithChatApplicationCommand registers a new discordgo.ChatApplicationCommand.
// This is syntactic sugar for WithApplicationCommand
func (r *Endpoint) WithChatApplicationCommand(name string, handler router.ApplicationCommandHandler) *Endpoint {
	return r.WithApplicationCommand(name, discordgo.ChatApplicationCommand, handler)
}

// WithUserApplicationCommand registers a new discordgo.UserApplicationCommand.
// This is syntactic sugar for WithApplicationCommand
func (r *Endpoint) WithUserApplicationCommand(name string, handler router.ApplicationCommandHandler) *Endpoint {
	return r.WithApplicationCommand(name, discordgo.UserApplicationCommand, handler)
}

// WithMessageApplicationCommand registers a new discordgo.MessageApplicationCommand.
// This is syntactic sugar for WithApplicationCommand
func (r *Endpoint) WithMessageApplicationCommand(name string, handler router.ApplicationCommandHandler) *Endpoint {
	return r.WithApplicationCommand(name, discordgo.MessageApplicationCommand, handler)
}

// WithApplicationCommand registers a new application command with the underlying Router.
func (r *Endpoint) WithApplicationCommand(name string, commandType discordgo.ApplicationCommandType, handler router.ApplicationCommandHandler) *Endpoint {
	r.router.RegisterCommand(name, commandType, handler)

	return r
}

// Handle handles the events.LambdaFunctionURLRequest.
// It should be registered to the Lambda Start in a function which is configured as a single-url function.
// See https://docs.aws.amazon.com/lambda/latest/dg/urls-configuration.html for more info.
func (r *Endpoint) Handle(ctx context.Context, event *events.LambdaFunctionURLRequest) (res *events.LambdaFunctionURLResponse, err error) {
	ctx, s := xray.BeginSubsegment(ctx, "handle")
	defer s.Close(err)
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	bs := []byte(event.Body)

	r.log.Info(
		"Received request",
		slog.String("user_agent", event.RequestContext.HTTP.UserAgent),
		slog.String("method", event.RequestContext.HTTP.Method),
		//slog.String("version", build.Version),
	)

	if err = r.verify(ctx, event); err != nil {
		r.log.Error("Failed to verify signature", "error", err)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	var i *discordgo.InteractionCreate
	if err = json.Unmarshal(bs, &i); err != nil {
		return nil, err
	}

	response, err := r.handleInteraction(ctx, i)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return &events.LambdaFunctionURLResponse{StatusCode: http.StatusAccepted}, nil
	}

	bs, err = json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Body:       string(bs),
	}, nil
}

// verify verifies the request using the ed25519 signature as per Discord's documentation.
// See https://discord.com/developers/docs/events/webhook-events#setting-up-an-endpoint-validating-security-request-headers.
func (r *Endpoint) verify(ctx context.Context, event *events.LambdaFunctionURLRequest) error {
	_, s := xray.BeginSubsegment(ctx, "verify")
	defer s.Close(nil)

	if len(r.publicKey) == 0 {
		return nil
	}

	headers := make(http.Header, len(event.Headers))
	for k, v := range event.Headers {
		headers.Add(k, v)
	}

	signature := headers.Get(headerSignature)
	if signature == "" {
		return errors.New("missing header X-Signature-Ed25519")
	}
	ts := headers.Get(headerTimestamp)
	if ts == "" {
		return errors.New("missing header X-Signature-Timestamp")
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	verify := append([]byte(ts), []byte(event.Body)...)

	if !ed25519.Verify(r.publicKey, verify, sig) {
		return errors.New("invalid signature")
	}

	return nil
}

// handleInteraction handles the discordgo.InteractionCreate, returning an optional sync response
func (r *Endpoint) handleInteraction(ctx context.Context, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	r.log.Info("Handling interaction", "type", i.Type, "interaction_id", i.ID)
	ctx, seg := xray.BeginSubsegment(ctx, "handle interaction")
	_ = seg.AddAnnotation("type", int(i.Type))
	defer seg.Close(nil)

	// build a session scoped for the interaction
	is := r.s
	if is == nil {
		is, _ = discordgo.New("Bot " + i.Token)
		is.Client = xray.Client(is.Client)
	}

	return r.router.HandleWithContext(ctx, is, i), nil
}
