package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rendau/fs/internal/adapters/httpapi/rest"
	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/spf13/viper"
)

func Execute() {
	var err error

	loadConf()

	debug := viper.GetBool("debug")

	app := struct {
		lg      *zap.St
		core    *core.St
		restApi *rest.St
	}{}

	app.lg, err = zap.New(viper.GetString("log_level"), debug, false)
	if err != nil {
		log.Fatal(err)
	}

	app.core = core.New(
		app.lg,
		viper.GetString("dir_path"),
	)

	app.restApi = rest.New(
		app.lg,
		viper.GetString("http_listen"),
		app.core,
	)

	app.lg.Infow(
		"Starting",
		"http_listen", viper.GetString("http_listen"),
	)

	app.restApi.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	var exitCode int

	select {
	case <-stop:
	case <-app.restApi.Wait():
		exitCode = 1
	}

	app.lg.Infow("Shutting down...")

	ctx, ctxCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer ctxCancel()

	err = app.restApi.Shutdown(ctx)
	if err != nil {
		app.lg.Errorw("Fail to shutdown http-api", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func loadConf() {
	viper.SetDefault("debug", "false")
	viper.SetDefault("http_listen", ":80")
	viper.SetDefault("log_level", "debug")

	confFilePath := os.Getenv("CONF_PATH")
	if confFilePath == "" {
		confFilePath = "conf.yml"
	}
	viper.SetConfigFile(confFilePath)
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	// viper.Set("dir_path", uriRPadSlash(viper.GetString("dir_path")))
}

func uriRPadSlash(uri string) string {
	if uri != "" && !strings.HasSuffix(uri, "/") {
		return uri + "/"
	}
	return uri
}
