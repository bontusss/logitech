package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePhoneRequiresE164(t *testing.T) {
	require.Equal(t, "+2348012345678", normalizePhone("+234 801-234-5678"))
	require.Empty(t, normalizePhone("08012345678"))
	require.Empty(t, normalizePhone("+0123456789"))
}
