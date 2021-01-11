package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const confPath = "test_conf.yml"
const dirPath = "test_dir"

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
		dirPath,
	)

	// Start tests
	code := m.Run()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	defer func() {
		_ = os.RemoveAll(filepath.Join(dirPath, "photos"))
	}()

	ctx := context.Background()

	const fileContent = "test_data"

	fileContentRaw := []byte(fileContent)

	fPath, err := app.core.Create(ctx, "photos", bytes.NewBuffer(fileContentRaw))
	require.Nil(t, err)

	fPathPrefix := "photos/" + time.Now().Format("2006/01/02") + "/"

	require.True(t, strings.HasPrefix(fPath, fPathPrefix))
	require.False(t, strings.Contains(strings.TrimPrefix(fPath, fPathPrefix), "/"))

	fContent, err := app.core.Get(ctx, fPath)
	require.Nil(t, err)
	require.Equal(t, fileContent, string(fContent))
}
