package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/bvieira/sv4git/v2/sv"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func configDefaultHandler() func(c *cli.Context) error {
	cfg := defaultConfig()
	return func(c *cli.Context) error {
		content, err := yaml.Marshal(&cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(content))
		return nil
	}
}

func configShowHandler(cfg Config) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		content, err := yaml.Marshal(&cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(content))
		return nil
	}
}

func currentVersionHandler(git sv.Git) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		lastTag := git.LastTag()

		currentVer, err := sv.ToVersion(lastTag)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from git tag, message: %v", lastTag, err)
		}
		fmt.Printf("%d.%d.%d\n", currentVer.Major(), currentVer.Minor(), currentVer.Patch())
		return nil
	}
}

func nextVersionHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		lastTag := git.LastTag()

		currentVer, err := sv.ToVersion(lastTag)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from git tag, message: %v", lastTag, err)
		}

		commits, err := git.Log(sv.NewLogRange(sv.TagRange, lastTag, ""))
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer, _ := semverProcessor.NextVersion(currentVer, commits)
		fmt.Printf("%d.%d.%d\n", nextVer.Major(), nextVer.Minor(), nextVer.Patch())
		return nil
	}
}

func commitLogHandler(git sv.Git) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		var commits []sv.GitCommitLog
		var err error
		tagFlag := c.String("t")
		rangeFlag := c.String("r")
		startFlag := c.String("s")
		endFlag := c.String("e")
		if tagFlag != "" && (rangeFlag != string(sv.TagRange) || startFlag != "" || endFlag != "") {
			return fmt.Errorf("cannot define tag flag with range, start or end flags")
		}

		if tagFlag != "" {
			commits, err = getTagCommits(git, tagFlag)
		} else {
			r, rerr := logRange(git, rangeFlag, startFlag, endFlag)
			if rerr != nil {
				return rerr
			}
			commits, err = git.Log(r)
		}
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		for _, commit := range commits {
			content, err := json.Marshal(commit)
			if err != nil {
				return err
			}
			fmt.Println(string(content))
		}
		return nil
	}
}

func getTagCommits(git sv.Git, tag string) ([]sv.GitCommitLog, error) {
	prev, _, err := getTags(git, tag)
	if err != nil {
		return nil, err
	}
	return git.Log(sv.NewLogRange(sv.TagRange, prev, tag))
}

func logRange(git sv.Git, rangeFlag, startFlag, endFlag string) (sv.LogRange, error) {
	switch rangeFlag {
	case string(sv.TagRange):
		return sv.NewLogRange(sv.TagRange, str(startFlag, git.LastTag()), endFlag), nil
	case string(sv.DateRange):
		return sv.NewLogRange(sv.DateRange, startFlag, endFlag), nil
	case string(sv.HashRange):
		return sv.NewLogRange(sv.HashRange, startFlag, endFlag), nil
	default:
		return sv.LogRange{}, fmt.Errorf("invalid range: %s, expected: %s, %s or %s", rangeFlag, sv.TagRange, sv.DateRange, sv.HashRange)
	}
}

func commitNotesHandler(git sv.Git, rnProcessor sv.ReleaseNoteProcessor, outputFormatter sv.OutputFormatter) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		var date time.Time

		rangeFlag := c.String("r")
		lr, err := logRange(git, rangeFlag, c.String("s"), c.String("e"))
		if err != nil {
			return err
		}

		commits, err := git.Log(lr)
		if err != nil {
			return fmt.Errorf("error getting git log from range: %s, message: %v", rangeFlag, err)
		}

		if len(commits) > 0 {
			date, _ = time.Parse("2006-01-02", commits[0].Date)
		}

		output, err := outputFormatter.FormatReleaseNote(rnProcessor.Create(nil, "", date, commits))
		if err != nil {
			return fmt.Errorf("could not format release notes, message: %v", err)
		}
		fmt.Println(output)
		return nil
	}
}

func releaseNotesHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, rnProcessor sv.ReleaseNoteProcessor, outputFormatter sv.OutputFormatter) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		var commits []sv.GitCommitLog
		var rnVersion *semver.Version
		var tag string
		var date time.Time
		var err error

		if tag = c.String("t"); tag != "" {
			rnVersion, date, commits, err = getTagVersionInfo(git, tag)
		} else {
			// TODO: should generate release notes if version was not updated?
			rnVersion, _, date, commits, err = getNextVersionInfo(git, semverProcessor)
		}

		if err != nil {
			return err
		}

		releasenote := rnProcessor.Create(rnVersion, tag, date, commits)
		output, err := outputFormatter.FormatReleaseNote(releasenote)
		if err != nil {
			return fmt.Errorf("could not format release notes, message: %v", err)
		}
		fmt.Println(output)
		return nil
	}
}

func getTagVersionInfo(git sv.Git, tag string) (*semver.Version, time.Time, []sv.GitCommitLog, error) {
	tagVersion, _ := sv.ToVersion(tag)

	previousTag, currentTag, err := getTags(git, tag)
	if err != nil {
		return nil, time.Time{}, nil, fmt.Errorf("error listing tags, message: %v", err)
	}

	commits, err := git.Log(sv.NewLogRange(sv.TagRange, previousTag, tag))
	if err != nil {
		return nil, time.Time{}, nil, fmt.Errorf("error getting git log from tag: %s, message: %v", tag, err)
	}

	return tagVersion, currentTag.Date, commits, nil
}

func getTags(git sv.Git, tag string) (string, sv.GitTag, error) {
	tags, err := git.Tags()
	if err != nil {
		return "", sv.GitTag{}, err
	}

	index := find(tag, tags)
	if index < 0 {
		return "", sv.GitTag{}, fmt.Errorf("tag: %s not found", tag)
	}

	previousTag := ""
	if index > 0 {
		previousTag = tags[index-1].Name
	}
	return previousTag, tags[index], nil
}

func find(tag string, tags []sv.GitTag) int {
	for i := 0; i < len(tags); i++ {
		if tag == tags[i].Name {
			return i
		}
	}
	return -1
}

func getNextVersionInfo(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) (*semver.Version, bool, time.Time, []sv.GitCommitLog, error) {
	lastTag := git.LastTag()

	commits, err := git.Log(sv.NewLogRange(sv.TagRange, lastTag, ""))
	if err != nil {
		return nil, false, time.Time{}, nil, fmt.Errorf("error getting git log, message: %v", err)
	}

	currentVer, _ := sv.ToVersion(lastTag)
	version, updated := semverProcessor.NextVersion(currentVer, commits)

	return version, updated, time.Now(), commits, nil
}

func tagHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		lastTag := git.LastTag()

		currentVer, err := sv.ToVersion(lastTag)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from git tag, message: %v", lastTag, err)
		}

		commits, err := git.Log(sv.NewLogRange(sv.TagRange, lastTag, ""))
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer, _ := semverProcessor.NextVersion(currentVer, commits)
		tagname, err := git.Tag(*nextVer)
		fmt.Println(tagname)
		if err != nil {
			return fmt.Errorf("error generating tag version: %s, message: %v", nextVer.String(), err)
		}
		return nil
	}
}

func getCommitType(cfg Config, p sv.MessageProcessor, input string) (string, error) {
	if input == "" {
		t, err := promptType(cfg.CommitMessage.Types)
		return t.Type, err
	}
	return input, p.ValidateType(input)
}

func getCommitScope(cfg Config, p sv.MessageProcessor, input string, noScope bool) (string, error) {
	if input == "" && !noScope {
		return promptScope(cfg.CommitMessage.Scope.Values)
	}
	return input, p.ValidateScope(input)
}

func getCommitDescription(p sv.MessageProcessor, input string) (string, error) {
	if input == "" {
		return promptSubject()
	}
	return input, p.ValidateDescription(input)
}

func getCommitBody(noBody bool) (string, error) {
	if noBody {
		return "", nil
	}

	var fullBody strings.Builder
	for body, err := promptBody(); body != "" || err != nil; body, err = promptBody() {
		if err != nil {
			return "", err
		}
		if fullBody.Len() > 0 {
			fullBody.WriteString("\n")
		}
		if body != "" {
			fullBody.WriteString(body)
		}
	}
	return fullBody.String(), nil
}

func getCommitIssue(cfg Config, p sv.MessageProcessor, branch string, noIssue bool) (string, error) {
	branchIssue, err := p.IssueID(branch)
	if err != nil {
		return "", err
	}

	if cfg.CommitMessage.IssueFooterConfig().Key == "" || cfg.CommitMessage.Issue.Regex == "" {
		return "", nil
	}

	if noIssue {
		return branchIssue, nil
	}

	return promptIssueID("issue id", cfg.CommitMessage.Issue.Regex, branchIssue)
}

func getCommitBreakingChange(noBreaking bool, input string) (string, error) {
	if noBreaking {
		return "", nil
	}

	if strings.TrimSpace(input) != "" {
		return input, nil
	}

	hasBreakingChanges, err := promptConfirm("has breaking change?")
	if err != nil {
		return "", err
	}
	if !hasBreakingChanges {
		return "", nil
	}

	return promptBreakingChanges()
}

func commitHandler(cfg Config, git sv.Git, messageProcessor sv.MessageProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		noBreaking := c.Bool("no-breaking")
		noBody := c.Bool("no-body")
		noIssue := c.Bool("no-issue")
		noScope := c.Bool("no-scope")
		inputType := c.String("type")
		inputScope := c.String("scope")
		inputDescription := c.String("description")
		inputBreakingChange := c.String("breaking-change")

		ctype, err := getCommitType(cfg, messageProcessor, inputType)
		if err != nil {
			return err
		}

		scope, err := getCommitScope(cfg, messageProcessor, inputScope, noScope)
		if err != nil {
			return err
		}

		subject, err := getCommitDescription(messageProcessor, inputDescription)
		if err != nil {
			return err
		}

		fullBody, err := getCommitBody(noBody)
		if err != nil {
			return err
		}

		issue, err := getCommitIssue(cfg, messageProcessor, git.Branch(), noIssue)
		if err != nil {
			return err
		}

		breakingChange, err := getCommitBreakingChange(noBreaking, inputBreakingChange)
		if err != nil {
			return err
		}

		header, body, footer := messageProcessor.Format(sv.NewCommitMessage(ctype, scope, subject, fullBody, issue, breakingChange))

		err = git.Commit(header, body, footer)
		if err != nil {
			return fmt.Errorf("error executing git commit, message: %v", err)
		}
		return nil
	}
}

func changelogHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, rnProcessor sv.ReleaseNoteProcessor, formatter sv.OutputFormatter) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		tags, err := git.Tags()
		if err != nil {
			return err
		}
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Date.After(tags[j].Date)
		})

		var releaseNotes []sv.ReleaseNote

		size := c.Int("size")
		all := c.Bool("all")
		addNextVersion := c.Bool("add-next-version")
		semanticVersionOnly := c.Bool("semantic-version-only")

		if addNextVersion {
			rnVersion, updated, date, commits, uerr := getNextVersionInfo(git, semverProcessor)
			if uerr != nil {
				return uerr
			}
			if updated {
				releaseNotes = append(releaseNotes, rnProcessor.Create(rnVersion, "", date, commits))
			}
		}
		for i, tag := range tags {
			if !all && i >= size {
				break
			}

			previousTag := ""
			if i+1 < len(tags) {
				previousTag = tags[i+1].Name
			}

			if semanticVersionOnly && !sv.IsValidVersion(tag.Name) {
				continue
			}

			commits, err := git.Log(sv.NewLogRange(sv.TagRange, previousTag, tag.Name))
			if err != nil {
				return fmt.Errorf("error getting git log from tag: %s, message: %v", tag.Name, err)
			}

			currentVer, _ := sv.ToVersion(tag.Name)
			releaseNotes = append(releaseNotes, rnProcessor.Create(currentVer, tag.Name, tag.Date, commits))
		}

		output, err := formatter.FormatChangelog(releaseNotes)
		if err != nil {
			return fmt.Errorf("could not format changelog, message: %v", err)
		}
		fmt.Println(output)

		return nil
	}
}

func validateCommitMessageHandler(git sv.Git, messageProcessor sv.MessageProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		branch := git.Branch()
		detached, derr := git.IsDetached()

		if messageProcessor.SkipBranch(branch, derr == nil && detached) {
			warnf("commit message validation skipped, branch in ignore list or detached...")
			return nil
		}

		if source := c.String("source"); source == "merge" {
			warnf("commit message validation skipped, ignoring source: %s...", source)
			return nil
		}

		filepath := filepath.Join(c.String("path"), c.String("file"))

		commitMessage, err := readFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read commit message, error: %s", err.Error())
		}

		if err := messageProcessor.Validate(commitMessage); err != nil {
			return fmt.Errorf("invalid commit message, error: %s", err.Error())
		}

		msg, err := messageProcessor.Enhance(branch, commitMessage)
		if err != nil {
			warnf("could not enhance commit message, %s", err.Error())
			return nil
		}
		if msg == "" {
			return nil
		}

		if err := appendOnFile(msg, filepath); err != nil {
			return fmt.Errorf("failed to append meta-informations on footer, error: %s", err.Error())
		}

		return nil
	}
}

func readFile(filepath string) (string, error) {
	f, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func appendOnFile(message, filepath string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(message)
	return err
}

func str(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
