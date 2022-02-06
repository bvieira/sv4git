package main

import (
	"testing"
)

func Test_checkTemplatesFiles(t *testing.T) {
	tests := []string{
		"resources/templates/changelog-md.tpl",
		"resources/templates/releasenotes-md.tpl",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			got, err := defaultTemplatesFS.ReadFile(tt)
			if err != nil {
				t.Errorf("missing template error = %v", err)
				return
			}
			if len(got) <= 0 {
				t.Errorf("empty template")
			}
		})
	}
}
