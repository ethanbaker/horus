package validation_test

import (
	"testing"

	"github.com/ethanbaker/horus/utils/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfirmation(t *testing.T) {
	assert := assert.New(t)

	assert.True(validation.ValidateConfirmation("yes"))
	assert.True(validation.ValidateConfirmation("yep"))
	assert.True(validation.ValidateConfirmation("yeah"))
	assert.True(validation.ValidateConfirmation("sure"))

	assert.False(validation.ValidateConfirmation("no"))
	assert.False(validation.ValidateConfirmation("nope"))
	assert.False(validation.ValidateConfirmation("never"))
	assert.False(validation.ValidateConfirmation("n"))
}

func TestValidateDenial(t *testing.T) {
	assert := assert.New(t)

	assert.False(validation.ValidateDenial("yes"))
	assert.False(validation.ValidateDenial("yep"))
	assert.False(validation.ValidateDenial("yeah"))
	assert.False(validation.ValidateDenial("sure"))

	assert.True(validation.ValidateDenial("no"))
	assert.True(validation.ValidateDenial("nope"))
	assert.True(validation.ValidateDenial("never"))
	assert.True(validation.ValidateDenial("n"))
}

func TestValidateStop(t *testing.T) {
	assert := assert.New(t)

	assert.True(validation.ValidateStop("stop"))
	assert.True(validation.ValidateStop("close"))
	assert.True(validation.ValidateStop("cancel"))
	assert.True(validation.ValidateStop("abort"))

	assert.False(validation.ValidateStop("continue"))
	assert.False(validation.ValidateStop("go"))
	assert.False(validation.ValidateStop("do it"))
	assert.False(validation.ValidateStop("commit"))
}
