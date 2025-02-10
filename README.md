# bot

A lightweight "framework" for building Discord bots in Go

## Usage

> [!TIP]
> Prefer a real example? Check out [elliotwms/pinbot](https://github.com/elliotwms/pinbot)

```go
package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/elliotwms/bot"
)

func main() {
	// initiate a session
	session, _ := discordgo.New("token")

	// build the bot
	b := bot.
		New("application_id", session).
		WithMigrationEnabled(true).
		WithHandler(readyEventHandler).
		WithApplicationCommand("foo", fooHandler)

	// set up a context
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	
	// build and run
	b.Build().Run(ctx)
}

```

## Features

Check out the `bot.Builder` for a full overview of capabilities, with some key ones listed below:

### Lifecycle Management

Start and run the bot with minimal code, stop the bot via context cancellation (`signal.NotifyContext` is your friend!). Provide an unconnected session and the package will manage opening and closing the session. 

### Configurable

Customise intents (`WithIntents`), enable optional logging (`WithLogger`), add discordgo event handlers (`WithHandler`) and more!

### Router

Instead of writing your own `InteractionCreate` handler, register application commands and handlers to a router. 

Enable deferred responses to have the router respond to the interaction with a deferred response, useful in scenarios where the interaction may take longer than the initial 3 seconds to complete

A custom router can be passed to the `bot` via `WithRouter`. See [interactions/router](/interactions/router) for more.

### Migrator

Migrate application commands on boot to reflect those registered with the bot. Provide a guild ID to create commands within the guild instead of globally (useful for testing).

### Health check

Enable an HTTP health check endpoint, which returns successfully when the bot is connected. Useful for running your bot in a containerised architecture