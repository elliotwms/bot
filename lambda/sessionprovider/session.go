package sessionprovider

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/bwmarrin/discordgo"
	"github.com/winebarrel/secretlamb"
)

type Provider func() (*discordgo.Session, error)

// ParamStore initialises the Discord Session using the token stored in param store
func ParamStore(ctx context.Context, paramName string) Provider {
	return func() (*discordgo.Session, error) {
		_, seg := xray.BeginSubsegment(ctx, "param store")
		defer seg.Close(nil)
		if paramName == "" {
			return nil, errors.New("empty discord token paramstore parameter name")
		}

		parameters := secretlamb.MustNewParameters()
		parameters.HTTPClient = xray.Client(parameters.HTTPClient)

		p, err := parameters.GetWithDecryption(paramName)
		if err != nil {
			return nil, err
		}

		if p == nil || p.Parameter.Value == "" {
			return nil, fmt.Errorf("parameter empty")
		}

		s, _ := discordgo.New("Bot " + p.Parameter.Value)
		s.Client = xray.Client(s.Client)

		return s, nil
	}
}

// Cached wraps a Provider, ensuring it is only called once
func Cached(f Provider) Provider {
	return sync.OnceValues(f)
}
