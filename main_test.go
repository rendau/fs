package main

import (
	"log"
	"os"
	"testing"

	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const confPath = "test_conf.yml"

var (
	app = struct {
		lg   *zap.St
		core *core.St
	}{}
)

func TestMain(m *testing.M) {
	var err error

	viper.SetConfigFile(confPath)
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	app.lg, err = zap.New(
		"info",
		true,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer app.lg.Sync()

	app.core = core.New(
		app.lg,
		"./test_dir/",
	)

	// Start tests
	code := m.Run()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	require.True(t, true, true)
}
