package main

import (
	pwl "github.com/justjanne/powerline-go/powerline"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func segmentJobs(p *powerline) []pwl.Segment {
	nJobs := -1

	ppid := os.Getppid()
	out, _ := exec.Command("ps", "-oppid=").Output()
	processes := strings.Split(string(out), "\n")
	for _, processPpidStr := range processes {
		currPpid := strings.TrimSpace(processPpidStr)
		processPpid, _ := strconv.ParseInt(currPpid, 10, 64)
		if int(processPpid) == ppid {
			nJobs++
		}
	}

	content := "\uf013"
	if nJobs > 1 {
		content = "\uf085"
	}
	if nJobs > 0 {
		return []pwl.Segment{{
			Name:       "jobs",
			Content:    content,
			Foreground: p.theme.JobsFg,
			Background: p.theme.JobsBg,
		}}
	}
	return []pwl.Segment{}
}
