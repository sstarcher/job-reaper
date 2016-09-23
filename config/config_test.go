package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSensuMap(t *testing.T) {
	var data = `
sensu:
  addr: localhost:3030
  templates:
    hello: bob
    yep: bob
stdout:
    level: info
`

	config := loadConfig([]byte(data))
	assert.Equal(t, config.Sensu.Templates["hello"], "bob", "they should be equal")
	assert.Equal(t, len(config.Sensu.Templates), 2, "they should be equal")

}

func TestBadMap(t *testing.T) {
	var data = `
sensu:
  addr: localhost:3030
  templates:
    hello: bob
    yep: bob
stdout:
    level: derp
`
	alerters := load([]byte(data))
	assert.Equal(t, len(*alerters), 0, "they should be equal")

}

func TestSingleAlerter(t *testing.T) {
	var data = `
stdout:
  level: info
`

	config := loadConfig([]byte(data))
	assert.Equal(t, config.Stdout.Level, "info", "they should be equal")
}

func TestInvalidLevel(t *testing.T) {
	var data = `
stdout:
  level: derp
`

	config := loadConfig([]byte(data))
	assert.Equal(t, config.Stdout.Level, "derp", "they should be equal")
}
