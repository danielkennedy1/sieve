package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	expectedTarget := "(a + (a * a) + (a * a * a) + (a * a * a * a))"
	expectedPopSize := 500
	expectedGenerations := 100

	cfg := DefaultConfig()

	t.Run("TopLevelDefaults", func(t *testing.T) {
		assert.Equal(t, expectedTarget, cfg.TargetExpressionString, "TargetExpressionString should match default")
		assert.Equal(t, expectedGenerations, cfg.Generations, "Generations should match default")
		assert.Equal(t, 1, cfg.NumVars, "NumVars should match default")
	})

	t.Run("NestedPopulationDefaults", func(t *testing.T) {
		assert.Equal(t, expectedPopSize, cfg.Population.Size, "Population.Size should match default")
		assert.Equal(t, 0.1, cfg.Population.MutationRate, "Population.MutationRate should match default")
		assert.Equal(t, 5, cfg.Population.MaxDepth, "Population.MaxDepth should match default")
	})
}

func TestLoadConfig_FromFileOverride(t *testing.T) {
	tomlContent := `
target_expression_string = "x + y"
generations = 500

[population]
size = 100
mutation_rate = 0.5
gene_length = 30
`
	originalWd, _ := os.Getwd()

	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "test.toml")
	err := os.WriteFile(tempFilePath, []byte(tomlContent), 0600)
	assert.NoError(t, err)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	defer func() {
		os.Chdir(originalWd)
	}()

	expectedTarget := "x + y"
	expectedGenerations := 500
	expectedNumSamples := 100
	expectedPopSize := 100
	expectedMutationRate := 0.5
	expectedMaxDepth := 5

	cfg, err := LoadConfig("test")

	assert.NoError(t, err, "LoadConfig should not return an error when file is present")
	assert.NotNil(t, cfg, "Config should not be nil")

	t.Run("TopLevelOverrides", func(t *testing.T) {
		assert.Equal(t, expectedTarget, cfg.TargetExpressionString, "TargetExpressionString should be overridden by file")
		assert.Equal(t, expectedGenerations, cfg.Generations, "Generations should be overridden by file")
		assert.Equal(t, expectedNumSamples, cfg.NumSamplesToGenerate, "NumSamplesToGenerate should use default")
	})

	t.Run("NestedPopulationOverrides", func(t *testing.T) {
		assert.Equal(t, expectedPopSize, cfg.Population.Size, "Population.Size should be overridden by file")
		assert.Equal(t, expectedMutationRate, cfg.Population.MutationRate, "Population.MutationRate should be overridden by file")
		assert.Equal(t, expectedMaxDepth, cfg.Population.MaxDepth, "Population.MaxDepth should use default")
		assert.Equal(t, 30, cfg.Population.GeneLength, "Population.GeneLength should be overridden by file")
	})
}
