package cmd

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/rendau/dop/adapters/client/httpc"
	"github.com/rendau/dop/adapters/client/httpc/httpclient"
	dopLoggerZap "github.com/rendau/dop/adapters/logger/zap"
	dopServerHttps "github.com/rendau/dop/adapters/server/https"
	"github.com/rendau/dop/dopTools"
	"github.com/rendau/fs/docs"
	"github.com/rendau/fs/internal/adapters/cleaner"
	cleanerCleaner "github.com/rendau/fs/internal/adapters/cleaner/cleaner"
	cleanerMock "github.com/rendau/fs/internal/adapters/cleaner/mock"
	"github.com/rendau/fs/internal/adapters/server/rest"
	"github.com/rendau/fs/internal/domain/core"
)

func Execute() {
	// var err error

	app := struct {
		lg         *dopLoggerZap.St
		cleaner    cleaner.Cleaner
		core       *core.St
		restApi    *rest.St
		restApiSrv *dopServerHttps.St
	}{}

	confLoad()
	confParse()

	app.lg = dopLoggerZap.New(conf.LogLevel, conf.Debug)

	if conf.CleanApiUrl != "" {
		app.cleaner = cleanerCleaner.New(
			app.lg,
			httpclient.New(app.lg, &httpc.OptionsSt{
				Client: &http.Client{
					Timeout:   time.Minute,
					Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
				},
				Uri:           conf.CleanApiUrl,
				Method:        "PUT",
				LogPrefix:     "Clean-api: ",
				RetryCount:    2,
				RetryInterval: 10 * time.Second,
			}),
		)
	} else {
		app.cleaner = cleanerMock.New()
	}

	app.core = core.New(
		app.lg,
		app.cleaner,
		conf.DirPath,
		conf.ImgMaxWidth,
		conf.ImgMaxHeight,
		conf.WmPath,
		conf.WmOpacity,
		conf.WmDirPathsParsed,
		conf.CacheCount,
		conf.CacheDuration,
		false,
	)

	docs.SwaggerInfo.Host = conf.SwagHost
	docs.SwaggerInfo.BasePath = conf.SwagBasePath
	docs.SwaggerInfo.Schemes = []string{conf.SwagSchema}
	docs.SwaggerInfo.Title = "FS service"

	// START

	app.lg.Infow("Starting")

	app.core.Start()

	app.restApiSrv = dopServerHttps.Start(
		conf.HttpListen,
		rest.GetHandler(
			app.lg,
			app.core,
			conf.HttpCors,
		),
		app.lg,
	)

	var exitCode int

	select {
	case <-dopTools.StopSignal():
	case <-app.restApiSrv.Wait():
		exitCode = 1
	}

	// STOP

	app.lg.Infow("Shutting down...")

	if !app.restApiSrv.Shutdown(20 * time.Second) {
		exitCode = 1
	}

	app.lg.Infow("Wait routines...")

	app.core.StopAndWaitJobs()

	app.lg.Infow("Exit")

	os.Exit(exitCode)
}
