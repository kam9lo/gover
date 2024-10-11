package prompt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// Selection implements selectable prompt item.
type Selection interface {
	Field() string
	Doc() string
}

type SelectionItem struct {
	Name string
	Help string
}

const (
	SelectTemplateItemActive   = `{{ "✔" | green | bold }} {{ .Name | cyan | bold }}` + SelectTemplateItemHelp
	SelectTemplateItemInactive = `  {{ .Name }}` + SelectTemplateItemHelp
	SelectTemplateItemHelp     = `{{ printf " (%s)" .Help | faint }}`
)

// Select displays prompt with selection. Name is the prefix displayed before
// selected item and options contains bulk of items to select from.
func Select[T Selection](name string, options []T) (string, error) {
	selectionItems := make([]SelectionItem, 0, len(options))
	for _, o := range options {
		selectionItems = append(selectionItems, SelectionItem{
			Name: o.Field(),
			Help: o.Doc(),
		})
	}

	prompt := promptui.Select{
		Label: name,
		Items: selectionItems,
		Templates: &promptui.SelectTemplates{
			Active:   SelectTemplateItemActive,
			Inactive: SelectTemplateItemInactive,
			Selected: SelectTemplateItemActive,
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("prompt run: %w", err)
	}

	return selectionItems[idx].Name, err
}

// TextInput displays prompt with simple text input where user provides any
// non-formatted text. Name is the prefix displayed before text input field.
func TextInput(name string, required bool) (string, error) {
	prompt := promptui.Prompt{
		Label: name,
		Validate: func(s string) error {
			if required && s == "" {
				return errors.New("required")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("prompt run: %w", err)
	}

	return wrapString(result, defaultTextInputWidth), nil
}

func wrapString(input string, width int) string {
	if len(input) <= width {
		return input
	}

	var result strings.Builder
	for len(input) > width {
		splitIndex := width
		for splitIndex > 0 && input[splitIndex] != ' ' {
			splitIndex--
		}

		if splitIndex == 0 {
			splitIndex = width
		}

		result.WriteString(input[:splitIndex] + "\n")

		input = strings.TrimSpace(input[splitIndex:])
	}

	result.WriteString(input)

	return result.String()
}

const (
	defaultTextInputWidth = 72
)
