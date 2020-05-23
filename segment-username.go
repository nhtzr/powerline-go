package main

import (
	pwl "github.com/justjanne/powerline-go/powerline"
	"os"
)

func segmentUser(p *powerline) []pwl.Segment {
	if shouldIgnore(p) {
		return []pwl.Segment{}
	}

	return []pwl.Segment{{
		Name:       "user",
		Content:    userPrompt(p),
		Foreground: fg(p),
		Background: bg(p),
	}}
}

func shouldIgnore(p *powerline) bool {
	if p.username == "" {
		return true
	}
	return p.username == *p.args.IgnoreUser
}

func fg(p *powerline) uint8 {
	return p.theme.UsernameFg
}

func bg(p *powerline) uint8 {
	var background uint8
	if os.Getuid() == 0 {
		background = p.theme.UsernameRootBg
	} else {
		background = p.theme.UsernameBg
	}
	return background
}

func userPrompt(p *powerline) string {
	if os.Geteuid() == 0 {
		return "\uf2bd"
	}
	var userPrompt string
	switch *p.args.Shell {
	case "bash":
		userPrompt = "\\u"
	case "zsh":
		userPrompt = "%n"
	default:
		userPrompt = p.username
	}
	return userPrompt
}
