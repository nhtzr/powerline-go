package main

import (
	"regexp"
	"strings"

	pwl "github.com/justjanne/powerline-go/powerline"
)

func segmentCwd(p *powerline) (segments []pwl.Segment) {
	cwd := p.cwd

	if strings.HasPrefix(cwd, p.userInfo.HomeDir) {
		cwd = "~" + cwd[len(p.userInfo.HomeDir):]
	}

	lencwd := len(cwd)
	if lencwd > 50 || (lencwd > TermWidth() * 40 / 100) {
		cwd = shorten(cwd)
	}

	segments = append(segments, pwl.Segment{
		Name:       "cwd",
		Content:    cwd,
		Foreground: p.theme.HomeFg,
		Background: p.theme.HomeBg,
	})

	return segments
}

func shorten(cwd string) string {
	if strings.HasPrefix(cwd, "/") {
		cwd = cwd[len("/"):]
	}
	dirs := strings.Split(cwd, "/")
	lastThreeIdx := len(dirs) - 3
	if lastThreeIdx > 0 {
		dirs = append([]string{"\uF141"}, dirs[lastThreeIdx:]...)
	}
	for i, dir := range dirs {
		if matches, _ := regexp.MatchString("^[0-9A-Za-z]{32}", dir); matches {
			dirs[i] = cap(dir, 9)
			continue
		}
		if matches, _ := regexp.MatchString("^[A-Za-z_ ]+$", dir); matches {
			dirs[i] = cap(dir, 3)
			continue
		}
		if matches, _ := regexp.MatchString("^\\.[A-Za-z_ ]+$", dir); matches {
			dirs[i] = cap(dir, 4)
			continue
		}

		dirs[i] = cap(dir, 9)
	}
	return strings.Join(dirs, "/")
}

func cap(dir string, maxLength int) string {
	i := min(maxLength, len(dir))
	return dir[:i]
}

func min(l int, r int) int {
	if l < r {
		return l
	}
	return r
}
