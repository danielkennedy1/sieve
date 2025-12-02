package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PopulationConfig struct {
	Size           int     `mapstructure:"size"`
	MutationRate   float64 `mapstructure:"mutation_rate"`
	CrossoverRate  float64 `mapstructure:"crossover_rate"`
	MaxDepth       int     `mapstructure:"max_depth"`
	GeneLength     int     `mapstructure:"gene_length"`
	TournamentSize int     `mapstructure:"tournament_size"`
	EliteCount     int     `mapstructure:"elite_count"`
}

type Config struct {
	// Sample Generation Settings (Top level)
	TargetExpressionString string `mapstructure:"target_expression_string"`
	NumSamplesToGenerate   int    `mapstructure:"num_samples_to_generate"`
	NumVars                int    `mapstructure:"num_vars"`

	// Population Settings (Nested)
	Population PopulationConfig `mapstructure:"population"`

	// Evolution Settings (Top level)
	Generations int `mapstructure:"generations"`

	// Fitness Settings (Top level)
	ParsiomonyPenalty float64 `mapstructure:"parsimony_penalty"`
	MaxGenes          int     `mapstructure:"max_genes"`

	// General Settings (Top level)
	BNFFilePath string `mapstructure:"bnf_file_path"`
}

func DefaultConfig() Config {
	return Config{
		TargetExpressionString: "(a + (a * a) + (a * a * a) + (a * a * a * a))",
		NumVars:                1,
		NumSamplesToGenerate:   100,

		// Initialize the nested PopulationConfig struct here
		Population: PopulationConfig{
			Size:          500,
			MutationRate:  0.1,
			CrossoverRate: 0.7,
			MaxDepth:      5,
		},

		Generations: 100,

		BNFFilePath: "data/lecture.bnf",
	}
}

func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()
	cfg := DefaultConfig()

	v.SetConfigName(configFile)
	v.SetConfigType("toml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		fmt.Println("Warning: Configuration file not found, using defaults and environment variables.")
	}

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// 3. Unmarshal into the Config struct
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return &cfg, nil
}
