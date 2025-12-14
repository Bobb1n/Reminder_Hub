package models_analyzer

import "strings"

type ArraysError struct {
	Errors []error
}

func (ae ArraysError) Error() string {
	if len(ae.Errors) == 0 {
		return ""
	}
	b := strings.Builder{}
	for i, err := range ae.Errors {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (ae *ArraysError) Append(err error) {
	if err != nil {
		return
	}
	ae.Errors = append(ae.Errors, err)
}
