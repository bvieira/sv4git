package main

import (
	"encoding/json"
	"fmt"
	"sv4git/sv"

	"github.com/urfave/cli"
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

		commits, err := git.Log(describe)
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
		describe := git.Describe()

		commits, err := git.Log(describe)
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

func releaseNotesHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, rnProcessor sv.ReleaseNoteProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		describe := git.Describe()

		currentVer, err := sv.ToVersion(describe)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from describe, message: %v", describe, err)
		}

		commits, err := git.Log(describe)
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer := semverProcessor.NextVersion(currentVer, commits)

		releasenote := rnProcessor.Get(commits)
		fmt.Println(rnProcessor.Format(releasenote, nextVer))
		return nil
	}
}

func tagHandler(git sv.Git, semverProcessor sv.SemVerCommitsProcessor, rnProcessor sv.ReleaseNoteProcessor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		describe := git.Describe()

		currentVer, err := sv.ToVersion(describe)
		if err != nil {
			return fmt.Errorf("error parsing version: %s from describe, message: %v", describe, err)
		}

		commits, err := git.Log(describe)
		if err != nil {
			return fmt.Errorf("error getting git log, message: %v", err)
		}

		nextVer := semverProcessor.NextVersion(currentVer, commits)
		fmt.Printf("%d.%d.%d\n", nextVer.Major(), nextVer.Minor(), nextVer.Patch())

		if err := git.Tag(nextVer); err != nil {
			return fmt.Errorf("error generate tag version: %s, message: %v", nextVer.String(), err)
		}
		return nil
	}
}
