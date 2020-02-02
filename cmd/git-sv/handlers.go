package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"sv4git/sv"
	"time"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
)

func currentVersionHandler(git sv.Git) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		describe := git.Describe()

		currentVer, err := sv.ToVersion(describe)
		if err != nil {
			return err
		}
		fmt.Printf("%d.%d.%d\n", currentVer.Major(), currentVer.Minor(), currentVer.Patch())
		return nil
	}
}

func nextVersionHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		describe := git.Describe()

		currentVer, err := sv.ToVersion(describe)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from describe, message: %v", describe, err)
		}

		commits, err := git.Log(describe, "")
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer := semverProcessor.NextVersion(currentVer, commits)
		fmt.Printf("%d.%d.%d\n", nextVer.Major(), nextVer.Minor(), nextVer.Patch())
		return nil
	}
}

func commitLogHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		var commits []sv.GitCommitLog
		var err error

		if tag := c.String("t"); tag != "" {
			commits, err = getTagCommits(git, tag)
		} else {
			commits, err = git.Log(git.Describe(), "")
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
	return git.Log(prev, tag)
}

func releaseNotesHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, rnProcessor sv.ReleaseNoteProcessor, outputFormatter sv.OutputFormatter) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		var commits []sv.GitCommitLog
		var rnVersion semver.Version
		var date time.Time
		var err error

		if tag := c.String("t"); tag != "" {
			rnVersion, date, commits, err = getTagVersionInfo(git, semverProcessor, tag)
		} else {
			rnVersion, date, commits, err = getNextVersionInfo(git, semverProcessor)
		}

		if err != nil {
			return err
		}

		releasenote := rnProcessor.Create(rnVersion, date, commits)
		fmt.Println(outputFormatter.FormatReleaseNote(releasenote))
		return nil
	}
}

func getTagVersionInfo(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, tag string) (semver.Version, time.Time, []sv.GitCommitLog, error) {
	tagVersion, err := sv.ToVersion(tag)
	if err != nil {
		return semver.Version{}, time.Time{}, nil, fmt.Errorf("error parsing version: %s from tag, message: %v", tag, err)
	}

	previousTag, currentTag, err := getTags(git, tag)
	if err != nil {
		return semver.Version{}, time.Time{}, nil, fmt.Errorf("error listing tags, message: %v", err)
	}

	commits, err := git.Log(previousTag, tag)
	if err != nil {
		return semver.Version{}, time.Time{}, nil, fmt.Errorf("error getting git log from tag: %s, message: %v", tag, err)
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

func getNextVersionInfo(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) (semver.Version, time.Time, []sv.GitCommitLog, error) {
	describe := git.Describe()

	currentVer, err := sv.ToVersion(describe)
	if err != nil {
		return semver.Version{}, time.Time{}, nil, fmt.Errorf("error parsing version: %s from describe, message: %v", describe, err)
	}

	commits, err := git.Log(describe, "")
	if err != nil {
		return semver.Version{}, time.Time{}, nil, fmt.Errorf("error getting git log, message: %v", err)
	}

	return semverProcessor.NextVersion(currentVer, commits), time.Now(), commits, nil
}

func tagHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		describe := git.Describe()

		currentVer, err := sv.ToVersion(describe)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from describe, message: %v", describe, err)
		}

		commits, err := git.Log(describe, "")
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer := semverProcessor.NextVersion(currentVer, commits)
		fmt.Printf("%d.%d.%d\n", nextVer.Major(), nextVer.Minor(), nextVer.Patch())

		if err := git.Tag(nextVer); err != nil {
			return fmt.Errorf("error generating tag version: %s, message: %v", nextVer.String(), err)
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
		for i, tag := range tags {
			if !all && i >= size {
				break
			}

			previousTag := ""
			if i+1 < len(tags) {
				previousTag = tags[i+1].Name
			}

			commits, err := git.Log(previousTag, tag.Name)
			if err != nil {
				return fmt.Errorf("error getting git log from tag: %s, message: %v", tag.Name, err)
			}

			currentVer, err := sv.ToVersion(tag.Name)
			if err != nil {
				return fmt.Errorf("error parsing version: %s from describe, message: %v", tag.Name, err)
			}
			releaseNotes = append(releaseNotes, rnProcessor.Create(currentVer, tag.Date, commits))
		}

		fmt.Println(formatter.FormatChangelog(releaseNotes))

		return nil
	}
}
