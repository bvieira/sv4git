package main

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/manifoldco/promptui"
)

type commitType struct {
	Type        string
	Description string
	Example     string
}

func promptType() (commitType, error) {
	items := []commitType{
		{Type: "build", Description: "changes that affect the build system or external dependencies", Example: "gradle, maven, go mod, npm"},
		{Type: "ci", Description: "changes to our CI configuration files and scripts", Example: "Circle, BrowserStack, SauceLabs"},
		{Type: "chore", Description: "update something without impacting the user", Example: "gitignore"},
		{Type: "docs", Description: "documentation only changes"},
		{Type: "feat", Description: "a new feature"},
		{Type: "fix", Description: "a bug fix"},
		{Type: "perf", Description: "a code change that improves performance"},
		{Type: "refactor", Description: "a code change that neither fixes a bug nor adds a feature"},
		{Type: "style", Description: "changes that do not affect the meaning of the code", Example: "white-space, formatting, missing semi-colons, etc"},
		{Type: "test", Description: "adding missing tests or correcting existing tests"},
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ .Type | white }} - {{ .Description | faint }}",
		Inactive: "  {{ .Type | white }} - {{ .Description | faint }}",
		Selected: `{{ "type:" | faint }} {{ .Type | white }}`,
		Details: `
{{ "Type:" | faint }}	{{ .Type }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "Example:" | faint }}	{{ .Example }}`,
	}

	i, err := promptSelect("type", items, template)
	if err != nil {
		return commitType{}, err
	}
	return items[i], nil
}

func promptScope(values []string) (string, error) {
	if len(values) > 0 {
		selected, err := promptSelect("scope", values, nil)
		if err != nil {
			return "", err
		}
		return values[selected], nil
	}
	return promptText("scope", "^[a-z0-9-]*$", "")
}

func promptSubject() (string, error) {
	return promptText("subject", "^[a-z].+$", "")
}

func promptBody() (string, error) {
	return promptText("body (leave empty to finish)", "^.*$", "")
}

func promptIssueID(issueLabel, issueRegex, defaultValue string) (string, error) {
	return promptText(issueLabel, "^("+issueRegex+")?$", defaultValue)
}

func promptBreakingChanges() (string, error) {
	return promptText("Breaking changes description", "[a-z].+", "")
}

func promptSelect(label string, items interface{}, template *promptui.SelectTemplates) (int, error) {
	if items == nil || reflect.TypeOf(items).Kind() != reflect.Slice {
		return 0, fmt.Errorf("items %v is not a slice", items)
	}

	prompt := promptui.Select{
		Label:     label,
		Size:      reflect.ValueOf(items).Len(),
		Items:     items,
		Templates: template,
	}

	index, _, err := prompt.Run()
	return index, err
}

func promptText(label, regex, defaultValue string) (string, error) {
	validate := func(input string) error {
		regex := regexp.MustCompile(regex)
		if !regex.MatchString(input) {
			return fmt.Errorf("invalid value, expected: %s", regex)
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	return prompt.Run()
}

func promptConfirm(label string) (bool, error) {
	r, err := promptText(label+" [y/n]", "^y|n$", "")
	if err != nil {
		return false, err
	}
	return r == "y", nil
}
