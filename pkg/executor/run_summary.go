package executor

import (
	"fmt"
	"strings"

	"github.com/junchaw/kubekraken/pkg/utils"
)

type RunSummary struct {
	Errors     []TaskResult `json:"errorTasks" yaml:"errors"`
	ErrorCount int          `json:"errorCount" yaml:"errorCount"`

	Warnings     []TaskResult `json:"warningTasks" yaml:"warnings"`
	WarningCount int          `json:"warningCount" yaml:"warningCount"`

	TotalCount int `json:"totalCount" yaml:"totalCount"`
}

func (s *RunSummary) ToText() string {
	text := "SUMMARY:\n"

	for _, result := range s.Errors {
		stderrStub := ""
		if len(result.Stderr) > 0 {
			stderrStub = ", stderr:"
		}
		text += fmt.Sprintf("- %s: error: %v%s\n", result.TaskItem.ID, result.Err, stderrStub)
		if len(result.Stderr) > 0 {
			text += strings.TrimSpace(string(result.Stderr)) + "\n"
		}
	}

	for _, result := range s.Warnings {
		text += fmt.Sprintf("- %s: stderr: %s\n", result.TaskItem.ID, strings.TrimSpace(string(result.Stderr)))
	}

	text += fmt.Sprintf("%d successful (%d with warnings), %d error, %d total\n",
		s.TotalCount-s.ErrorCount,
		s.WarningCount,
		s.ErrorCount,
		s.TotalCount,
	)

	return text
}

// ToStyledText returns a slice of StyleText for the summary,
// we don't return a string because there is some weird issue with line breaks when joining styled text together.
func (s *RunSummary) ToStyledText() []utils.StyleText {
	summaryLines := []utils.StyleText{
		{Text: "SUMMARY:", Style: &utils.Style.Text},
	}

	errClusters := map[string]bool{} // we keep this map to avoid duplicated warning messages

	for _, result := range s.Errors {
		errClusters[result.TaskItem.ID] = true

		stderrStub := ""
		if len(result.Stderr) > 0 {
			stderrStub = ", stderr:"
		}
		summaryLines = append(summaryLines, utils.StyleText{
			Text:  fmt.Sprintf("- %s: error: %v%s", result.TaskItem.ID, result.Err, stderrStub),
			Style: &utils.Style.Warning,
		})
		if len(result.Stderr) > 0 {
			summaryLines = append(summaryLines, utils.StyleText{
				Text:  strings.TrimSpace(string(result.Stderr)),
				Style: &utils.Style.Warning,
			})
		}
	}

	for _, result := range s.Warnings {
		if errClusters[result.TaskItem.ID] { // the warning should already be printed in the error section
			continue
		}
		summaryLines = append(summaryLines, utils.StyleText{
			Text:  fmt.Sprintf("- %s: stderr: %s", result.TaskItem.ID, strings.TrimSpace(string(result.Stderr))),
			Style: &utils.Style.Warning,
		})
	}

	summaryLines = append(summaryLines, utils.StyleText{
		Text: fmt.Sprintf("%d successful (%d with warnings), %d error, %d total",
			s.TotalCount-s.ErrorCount,
			s.WarningCount,
			s.ErrorCount,
			s.TotalCount,
		),
		Style: &utils.Style.Text,
	})

	return summaryLines
}
