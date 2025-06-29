package main

import (
	"fmt"
	"os"
	"strings"
	"xrf197ilz35aq/internal"
)

func main() {
	env := getAppEnv()
	config, err := internal.GetConfig(strings.ToLower(env))
	if err != nil {
		fmt.Println("failed to load config:", err)
		return
	}

	logger, err := internal.SetupLogger(env, config.Log)
	if err != nil {
		fmt.Printf("Failed to setup logger: %v\n", err)
		return
	}

	logger.Info("***** xrf197ilz35aq started successfully *******")
}

func getAppEnv() string {
	env, ok := os.LookupEnv(internal.Environment)
	if !ok || env == "" {
		env = internal.DevelopEnv
	}

	switch env {
	case internal.StagingEnv:
		return "STAGING"
	case internal.ProductionEnv, internal.LiveEnv:
		return internal.LiveEnv
	default:
		return internal.DevelopEnv
	}
}
