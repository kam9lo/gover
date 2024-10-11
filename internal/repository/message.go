package repository

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	templateParamsRgx    = regexp.MustCompile(`{{\.\w+}}`)
	templateParamNameRgx = regexp.MustCompile(`\w+`)
)

// Message is a key-value store of commit message parameters, where key is the
// name of provided in template field and value - matching pattern text.
type Message map[string]string

// ParseMessage returns message with parsed from template and commit message
// parameters.
func ParseMessage(template, message string, required ...string) (Message, error) {
	tmplParams := map[string]string{}
	bParams := templateParamsRgx.FindAllStringIndex(template, -1)
	cursor := 0

	for i, bp := range bParams {
		if len(message[cursor:]) == 0 {
			break
		}

		ph := placeholder{
			beginIdx: bp[0],
			endIdx:   bp[1],
		}
		paramName := templateParamNameRgx.FindString(
			template[ph.beginIdx:ph.endIdx],
		)

		ph.sepBeginIdx = ph.endIdx
		if next, ok := next(bParams, i); ok {
			ph.sepEndIdx = next[0]
		} else {
			ph.sepEndIdx = len(template)
		}

		phraseSep := template[ph.sepBeginIdx:ph.sepEndIdx]
		if len(phraseSep) == 0 {
			tmplParams[paramName] = message[cursor:]
			break
		}

		phraseEndIdx := strings.Index(message[cursor:], phraseSep)
		if phraseEndIdx < 0 {
			tmplParams[paramName] = message[cursor:]
			break
		}

		phrase := message[cursor : cursor+phraseEndIdx]
		tmplParams[paramName] = phrase
		cursor += len(phrase) + len(phraseSep)
	}

	for _, r := range required {
		if tmplParams[r] == "" {
			return nil, fmt.Errorf(
				missingRequiredErrMsg,
				template,
				strings.Join(required, ","),
				r,
			)
		}
	}

	return tmplParams, nil
}

type placeholder struct {
	beginIdx    int
	endIdx      int
	sepBeginIdx int
	sepEndIdx   int
}

func isIndexLast[T any](slice []T, idx int) bool {
	return idx == len(slice)-1
}

func next[T any](slice []T, idx int) (ret T, ok bool) {
	if isIndexLast(slice, idx) {
		return
	}
	return slice[idx+1], true
}

const missingRequiredErrMsg = `
missing required message parameter:
-----------------------------------
template:
%s
-----------------------------------
required:
%s
-----------------------------------
missing: %s`
