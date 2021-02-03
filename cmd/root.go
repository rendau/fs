package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rendau/fs/internal/adapters/cleaner/cleaner"
	cleanerMock "github.com/rendau/fs/internal/adapters/cleaner/mock"
	"github.com/rendau/fs/internal/adapters/httpapi/rest"
	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/rendau/fs/internal/interfaces"
	"github.com/spf13/viper"
)

func Execute() {
	var err error

	loadConf()

	fDebug := viper.GetBool("debug")

	app := struct {
		lg      interfaces.Logger
		cleaner interfaces.Cleaner
		core    *core.St
		restApi *rest.St
	}{}

	app.lg, err = zap.New(viper.GetString("log_level"), fDebug, false)
	if err != nil {
		log.Fatal(err)
	}

	httpListen := viper.GetString("http_listen")
	dirPath := viper.GetString("dir_path")
	cleanApiUrl := viper.GetString("clean_api_url")
	imgMaxWidth := viper.GetInt("img_max_width")
	imgMaxHeight := viper.GetInt("img_max_height")
	wmPath := viper.GetString("wm_path")
	wmOpacity := viper.GetFloat64("wm_opacity")
	wmDirPaths := viper.GetString("wm_dir_paths")
	cacheCount := viper.GetInt("cache_count")
	cacheDuration := viper.GetDuration("cache_duration")

	if cleanApiUrl != "" {
		app.cleaner = cleaner.New(app.lg, cleanApiUrl)
	} else {
		app.cleaner = cleanerMock.New()
	}

	app.core = core.New(
		app.lg,
		dirPath,
		imgMaxWidth,
		imgMaxHeight,
		wmPath,
		wmOpacity,
		parseWMarkDirPaths(wmDirPaths),
		app.cleaner,
		cacheCount,
		cacheDuration,
		false,
	)

	app.restApi = rest.New(
		app.lg,
		httpListen,
		app.core,
	)

	app.lg.Infow(
		"Starting",
		"http_listen", httpListen,
		"dir_path", dirPath,
		"clean_api_url", cleanApiUrl,
		"img_max_width", imgMaxWidth,
		"img_max_height", imgMaxHeight,
		"wm_path", wmPath,
		"wm_opacity", wmOpacity,
		"wm_dir_paths", wmDirPaths,
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

	app.lg.Infow("Wait routines...")

	app.core.StopAndWaitJobs()

	os.Exit(exitCode)
}

func parseWMarkDirPaths(src string) []string {
	result := make([]string, 0)

	for _, p := range strings.Split(viper.GetString("wm_dir_paths"), ";") {
		if p != "" {
			result = append(result, p)
		}
	}

	return result
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
}
