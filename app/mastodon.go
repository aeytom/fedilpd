package app

import (
	"context"
	"log"

	"github.com/mattn/go-mastodon"
)

type MastodonConfig struct {
	// https://botsin.space
	Server string `yaml:"server,omitempty"`
	// Client key: kZoi323…
	ClientID string `yaml:"client_id,omitempty"`
	// Client secret: ose…
	ClientSecret string `yaml:"client_secret,omitempty"`
	// Application name: fedilpd
	ClientName string `yaml:"client_name,omitempty"`
	// Scopes: read write follow
	Scopes string `yaml:"scopes,omitempty"`
	// Application website: https://berlin.de/presse
	Website string `yaml:"website,omitempty"`
	// Redirect URI: urn:ietf:wg:oauth:2.0:oob
	RedirectURI string `yaml:"redirect_uri,omitempty"`
	// Your access token: Rdn…
	Token string `yaml:"token,omitempty"`
	//
	UserAgent string `yaml:"user_agent,omitempty"`
}

const (
	UserAgent = "fedibde/0.01 (2023-12-25)"
)

func (s *Settings) GetClient() *mastodon.Client {
	if s.Mastodon.UserAgent == "" {
		s.Mastodon.UserAgent = UserAgent
	}
	c := &mastodon.Client{
		Config: &mastodon.Config{
			Server:       s.Mastodon.Server,
			ClientID:     s.Mastodon.ClientID,
			ClientSecret: s.Mastodon.ClientSecret,
			AccessToken:  s.Mastodon.Token,
		},
		UserAgent: s.Mastodon.UserAgent,
	}
	return c
}

func (s *Settings) GetApp() {
	if s.Mastodon.Scopes == "" {
		s.Mastodon.Scopes = "read write follow"
	}
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     s.Mastodon.Server,
		ClientName: s.Mastodon.ClientName,
		Scopes:     s.Mastodon.Scopes,
		Website:    s.Mastodon.Website,
	})
	if err != nil {
		log.Fatal(err)
	}

	s.Logf("client-id: %s\n", app.ClientID)
	s.Logf("client-secret: %s\n", app.ClientSecret)
}
