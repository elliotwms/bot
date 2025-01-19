package bot

import (
	"testing"

	"github.com/elliotwms/fakediscord/pkg/fakediscord"
)

const appID = "1881067102262001664"

func TestMain(m *testing.M) {
	fakediscord.Configure("http://localhost:8080/")
	m.Run()
}
