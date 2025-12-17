package models_analyzer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArraysError_Error_Empty(t *testing.T) {
	ae := &ArraysError{}
	assert.Equal(t, "", ae.Error())
}

func TestArraysError_Error_SingleError(t *testing.T) {
	ae := &ArraysError{
		Errors: []error{errors.New("error1")},
	}
	assert.Equal(t, "error1", ae.Error())
}

func TestArraysError_Error_MultipleErrors(t *testing.T) {
	ae := &ArraysError{
		Errors: []error{
			errors.New("error1"),
			errors.New("error2"),
			errors.New("error3"),
		},
	}
	result := ae.Error()
	assert.Contains(t, result, "error1")
	assert.Contains(t, result, "error2")
	assert.Contains(t, result, "error3")
	assert.Contains(t, result, "; ")
}

func TestArraysError_Append_WithError(t *testing.T) {
	ae := &ArraysError{}
	ae.Append(errors.New("test error"))
	
	assert.Len(t, ae.Errors, 1)
	assert.Equal(t, "test error", ae.Errors[0].Error())
}

func TestArraysError_Append_WithNil(t *testing.T) {
	ae := &ArraysError{}
	ae.Append(nil)
	
	assert.Len(t, ae.Errors, 0)
}

func TestArraysError_Append_Multiple(t *testing.T) {
	ae := &ArraysError{}
	ae.Append(errors.New("error1"))
	ae.Append(errors.New("error2"))
	ae.Append(nil)
	ae.Append(errors.New("error3"))
	
	assert.Len(t, ae.Errors, 3)
}

