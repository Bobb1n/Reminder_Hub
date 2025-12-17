package aiagent

import (
	"testing"

	"reminder-hub/services/analyzer/internal/ai_agent/mistral"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	mistralAgent := &mistral.MistralAgent{}
	agent := NewAgent(mistralAgent)

	assert.NotNil(t, agent)
	assert.Equal(t, mistralAgent, agent.mistralAgent)
}

func TestNewAgent_WithNil(t *testing.T) {
	agent := NewAgent(nil)

	assert.NotNil(t, agent)
	assert.Nil(t, agent.mistralAgent)
}
