package connector

import (
	"testing"
)

func TestLoadManifest(t *testing.T) {
	m := LoadManifest()
	if m == nil {
		t.Errorf("no such file or directory")
	}

	if m.Name != "ethereum" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Name, "ethereum")
	}
	if m.Author != "nakji" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Author, "nakji")
	}
	if m.Version.String() != "0.0.0" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Version, "0.0.0")
	}
	if m.Url != "https://nakji.network/" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Url, "https://nakji.network/")
	}
	if m.PrimaryColor != "#1c1d26" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.PrimaryColor, "#1c1d26")
	}
	if m.Links.Github != "https://github.com/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Github, "https://github.com/nakji-network")
	}
	if m.Links.Twitter != "https://twitter.com/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Twitter, "https://twitter.com/nakji-network")
	}
	if m.Links.Discord != "https://discord.gg/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Discord, "https://discord.gg/nakji-network")
	}
	if m.Links.Telegram != "https://t.me/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Telegram, "https://t.me/nakji-network")
	}
	if m.Links.Medium != "https://medium.com/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Medium, "https://medium.com/nakji-network")
	}
}

func TestNewManifest(t *testing.T) {
	urlOption := WithUrl("https://nakji.network/")

	linksOptions := WithLinks(link{
		Github:  "https://github.com/nakji-network",
		Discord: "https://discord.gg/nakji-network",
	})

	m := NewManifest("ethereum", "nakji", "0.1.2", urlOption, linksOptions)

	if m.Name != "ethereum" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Name, "ethereum")
	}
	if m.Author != "nakji" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Author, "nakji")
	}
	if m.Version.String() != "0.1.2" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Version, "0.1.2")
	}
	if m.Url != "https://nakji.network/" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Url, "https://nakji.network/")
	}
	if m.PrimaryColor != "" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.PrimaryColor, "")
	}
	if m.Links.Github != "https://github.com/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Github, "https://github.com/nakji-network")
	}
	if m.Links.Twitter != "" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Twitter, "")
	}
	if m.Links.Discord != "https://discord.gg/nakji-network" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Discord, "https://discord.gg/nakji-network")
	}
	if m.Links.Telegram != "" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Telegram, "")
	}
	if m.Links.Medium != "" {
		t.Errorf("Error loading manifest.yaml: got=%q want=%q", m.Links.Medium, "")
	}
}
