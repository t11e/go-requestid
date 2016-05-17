package requestid_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t11e/go-requestid"
)

func TestConfig_MakeID(t *testing.T) {
	c := requestid.Config{}
	id, err := c.MakeID()
	assert.NoError(t, err)
	assert.Equal(t, 16, len(id))

	id2, err := c.MakeID()
	assert.NoError(t, err)
	assert.NotEqual(t, id, id2)
}

func TestConfig_SetGenerator(t *testing.T) {
	idx := 0
	c := requestid.Config{Generator: func() (string, error) {
		idx += 1
		return fmt.Sprintf("%d", idx), nil
	}}
	id, err := c.MakeID()
	assert.NoError(t, err)
	assert.Equal(t, "1", id)

	id2, err := c.MakeID()
	assert.NoError(t, err)
	assert.Equal(t, "2", id2)
}

func TestConfig_SetGenerator_Error(t *testing.T) {
	c := requestid.Config{Generator: func() (string, error) {
		return "", errors.New("test error")
	}}
	id, err := c.MakeID()
	assert.Equal(t, errors.New("test error"), err)
	assert.Equal(t, "", id)
}
