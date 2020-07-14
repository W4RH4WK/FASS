package fass

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCourseUserAuthorization(t *testing.T) {
	course := Course{
		Users: []Token{"T0K3N"},
	}

	assert.True(t, course.IsUserAuthorized("T0K3N"))
	assert.False(t, course.IsUserAuthorized("C0FF33"))
}
