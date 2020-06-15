package main

import (
	"fmt"
	pwl "github.com/justjanne/powerline-go/powerline"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type repoStats struct {
	ahead      int
	behind     int
	untracked  int
	notStaged  int
	staged     int
	conflicted int
	stashed    int
}

func (r repoStats) dirty() bool {
	return r.untracked+r.notStaged+r.staged+r.conflicted > 0
}

func addRepoStatsSegment(content string, nChanges int, symbol string) string {
	if nChanges <= 0 {
		return content
	}
	if content != "" {
		content = content + " "
	}
	return fmt.Sprintf("%s%d%s", content, nChanges, symbol)
}

func (r repoStats) GitSegments(p *powerline) []pwl.Segment {
	var content string
	content = addRepoStatsSegment(content, r.ahead, p.symbolTemplates.RepoAhead)
	content = addRepoStatsSegment(content, r.behind, p.symbolTemplates.RepoBehind)
	content = addRepoStatsSegment(content, r.stashed, p.symbolTemplates.RepoBehind)
	if content == "" {
		return []pwl.Segment{}
	}
	return []pwl.Segment{{
		Name:       "git-status",
		Content:    content,
		Foreground: p.theme.GitAheadFg,
		Background: p.theme.GitAheadBg,
	}}
}

var branchRegex = regexp.MustCompile(`^## (?P<local>\S+?)(\.{3}(?P<remote>\S+?)( \[(ahead (?P<ahead>\d+)(, )?)?(behind (?P<behind>\d+))?])?)?$`)

func groupDict(pattern *regexp.Regexp, haystack string) map[string]string {
	match := pattern.FindStringSubmatch(haystack)
	result := make(map[string]string)
	if len(match) > 0 {
		for i, name := range pattern.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
	}
	return result
}

func gitProcessEnv() []string {
	home, _ := os.LookupEnv("HOME")
	path, _ := os.LookupEnv("PATH")
	env := map[string]string{
		"LANG": "C",
		"HOME": home,
		"PATH": path,
	}
	result := make([]string, 0)
	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}
	return result
}

func runGitCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	command.Env = gitProcessEnv()
	out, err := command.Output()
	return string(out), err
}

func parseGitBranchInfo(status []string) map[string]string {
	return groupDict(branchRegex, status[0])
}

func getGitDetachedBranch(p *powerline) string {
	out, err := runGitCommand("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		out, err := runGitCommand("git", "symbolic-ref", "--short", "HEAD")
		if err != nil {
			return "Error"
		}
		return strings.SplitN(out, "\n", 2)[0]
	}
	detachedRef := strings.SplitN(out, "\n", 2)
	return fmt.Sprintf("%s %s", p.symbolTemplates.RepoDetached, detachedRef[0])
}

func parseGitStats(status []string) repoStats {
	stats := repoStats{}
	if len(status) > 1 {
		for _, line := range status[1:] {
			if len(line) > 2 {
				code := line[:2]
				switch code {
				case "??":
					stats.untracked++
				case "DD", "AU", "UD", "UA", "DU", "AA", "UU":
					stats.conflicted++
				default:
					if code[0] != ' ' {
						stats.staged++
					}

					if code[1] != ' ' {
						stats.notStaged++
					}
				}
			}
		}
	}
	return stats
}

func repoRoot(path string) (string, error) {
	out, err := runGitCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func indexSize(root string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(root, ".git", "index"))
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func segmentGit(p *powerline) []pwl.Segment {
	repoRoot, err := repoRoot(p.cwd)
	if err != nil {
		return []pwl.Segment{}
	}

	if len(p.ignoreRepos) > 0 && p.ignoreRepos[repoRoot] {
		return []pwl.Segment{}
	}

	status, err := status(p)
	if err != nil {
		return []pwl.Segment{}
	}

	branchInfo := parseGitBranchInfo(status)
	stats := genStats(status, branchInfo)
	branch := genBranch(p, branchInfo, stats)

	var foreground, background uint8
	if dirty() {
		foreground = p.theme.RepoDirtyFg
		background = p.theme.RepoDirtyBg
	} else {
		foreground = p.theme.RepoCleanFg
		background = p.theme.RepoCleanBg
	}
	segments := []pwl.Segment{{
		Name:       "git-branch",
		Content:    branch,
		Foreground: foreground,
		Background: background,
	}}
	return append(segments, stats.GitSegments(p)...)
}

func status(p *powerline) ([]string, error) {
	indexSize, err := indexSize(p.cwd)
	args := []string{
		"status", "--porcelain", "-b", "--ignore-submodules",
	}
	if *p.args.GitAssumeUnchangedSize > 0 && indexSize > (*p.args.GitAssumeUnchangedSize*1024) {
		args = append(args, "-uno")
	}
	out, err := runGitCommand("git", args...)
	if err != nil {
		return nil, err
	}
	status := strings.Split(out, "\n")
	return status, nil
}

func genBranch(p *powerline, branchInfo map[string]string, stats repoStats) string {
	var branch string
	if branchInfo["local"] != "" {
		branch = branchInfo["local"]
	} else {
		branch = getGitDetachedBranch(p)
	}

	lenbranch := len(branch)
	if lenbranch > 14 || (lenbranch > TermWidth()*25/100) {
		branch = shortenBranch(branch)
	}
	branch = fmt.Sprintf("%s%s", p.symbolTemplates.RepoBranch, branch)
	branch = stats.appendRepostatSymbols(p, branch)
	branch = strings.TrimSpace(branch)
	return branch
}

func genStats(status []string, branchInfo map[string]string) repoStats {
	stats := parseGitStats(status)
	if branchInfo["local"] != "" {
		ahead, _ := strconv.ParseInt(branchInfo["ahead"], 10, 32)
		stats.ahead = int(ahead)

		behind, _ := strconv.ParseInt(branchInfo["behind"], 10, 32)
		stats.behind = int(behind)
	}

	out, err := runGitCommand("git", "rev-list", "-g", "refs/stash")
	if err == nil && len(out) > 0 {
		stats.stashed = len(strings.Split(out, "\n")) - 1
	}
	return stats
}

func shortenPrefix(branch string, prefix string, to string) string {
	if strings.HasPrefix(branch, prefix) {
		return fmt.Sprintf("%s%s", to, branch[len(prefix):])
	}
	return branch
}

func shortenBranch(branch string) string {
	branch = shortenPrefix(branch, "release/", "r/")
	branch = shortenPrefix(branch, "release-", "r/")
	branch = shortenPrefix(branch, "feature/", "f/")
	branch = shortenPrefix(branch, "feat/", "f/")
	branch = shortenPrefix(branch, "chore/", "c/")
	branch = shortenPrefix(branch, "master", "m")
	return branch
}

func appendRepostatSymbol(branch string, nChanges int, symbol string, pre string) (string, string) {
	if nChanges <= 0 {
		return pre, branch
	}
	return "", fmt.Sprintf("%s%s%s", branch, pre, symbol)
}

func (r repoStats) appendRepostatSymbols(p *powerline, branch string) string {
	pre := " "
	pre, branch = appendRepostatSymbol(branch, r.notStaged, p.symbolTemplates.RepoNotStaged, pre)
	pre, branch = appendRepostatSymbol(branch, r.conflicted, p.symbolTemplates.RepoConflicted, pre)
	return branch
}

func dirty() bool {
	describe, err := runGitCommand("git", "describe", "--always", "--dirty")
	if err != nil {
		return true
	}
	describe = strings.TrimSpace(describe)
	return strings.HasSuffix(describe, "-dirty")
}
