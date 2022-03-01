package sv

import "time"

func timeFormat(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(format)
}

func getSection(sections []ReleaseNoteSection, name string) ReleaseNoteSection {
	for _, section := range sections {
		if section.SectionName() == name {
			return section
		}
	}
	return nil
}
