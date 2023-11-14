package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanerConfig_DefaultVerboseIsFalse(t *testing.T) {
	c := NewCleanerConfig()
	c.profile = "some-profile"

	assert.Equal(t, false, c.verbose)
}

func TestCleanerConfig_DefaultProcessAllIsFalse(t *testing.T) {
	c := NewCleanerConfig()
	c.profile = "some-profile"

	assert.Equal(t, false, c.processAll)
}

func TestCleanerConfig_ValidateProfileIsSet(t *testing.T) {
	c := NewCleanerConfig()
	c.profile = "some-profile"
	c.stackToClean = "some-stack"

	assert.NoError(t, c.validate())
}

func TestCleanerConfig_ValidateProcessAllIsTrueIfStackAll(t *testing.T) {
	c := NewCleanerConfig()
	c.stackToClean = "all"
	c.profile = "some-profile"
	c.validate()

	assert.Equal(t, true, c.processAll)
}
