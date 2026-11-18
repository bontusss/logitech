package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductionRequiresPlunkCredentials(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("x", 40))
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("PLUNK_SECRET_KEY", "")
	t.Setenv("PLUNK_FROM_EMAIL", "")

	_, err := Load()
	require.ErrorContains(t, err, "PLUNK_SECRET_KEY")

	t.Setenv("PLUNK_SECRET_KEY", "sk_test")
	_, err = Load()
	require.ErrorContains(t, err, "PLUNK_FROM_EMAIL")

	t.Setenv("PLUNK_FROM_EMAIL", "hello@example.com")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "sk_test", cfg.PlunkSecretKey)
	require.Equal(t, "hello@example.com", cfg.PlunkFromEmail)
}
