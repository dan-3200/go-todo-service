package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"app/internal/infrastructure/config"
	"app/internal/infrastructure/telemetry"
)

func TestSetupTelemetrySkipsInitializationWithoutCollector(t *testing.T) {
	cleanup, err := telemetry.SetupTelemetry(context.Background(), config.TelemetryConfig{})

	require.NoError(t, err)
	require.Nil(t, cleanup)
}

func TestSetTelemetryRejectsEmptyCollector(t *testing.T) {
	cleanup, err := telemetry.SetTelemetry(context.Background(), telemetry.TelemetryConfig{})

	require.Error(t, err)
	require.Nil(t, cleanup)
	require.Equal(t, "endereço do collector não definido", err.Error())
}
