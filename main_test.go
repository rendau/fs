package main

import (
	"bytes"
	"context"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rendau/fs/internal/domain/entities"

	"github.com/disintegration/imaging"

	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const confPath = "test_conf.yml"
const dirPath = "test_dir"
const imgMaxWidth = 1000
const imgMaxHeight = 1000

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
		imgMaxWidth,
		imgMaxHeight,
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

	fileContentRaw := []byte("test_data")

	fPath, err := app.core.Create(ctx, "photos", "data.txt", bytes.NewBuffer(fileContentRaw))
	require.Nil(t, err)

	fPathPrefix := "photos/" + time.Now().Format("2006/01/02") + "/"

	require.True(t, strings.HasPrefix(fPath, fPathPrefix))
	require.False(t, strings.Contains(strings.TrimPrefix(fPath, fPathPrefix), "/"))

	fName, fContent, err := app.core.Get(ctx, fPath, &entities.ImgParsSt{})
	require.Nil(t, err)
	require.NotNil(t, fContent)
	require.Equal(t, "test_data", string(fContent))
	require.NotEmpty(t, fName)

	largeImg := imaging.New(imgMaxWidth+500, imgMaxHeight+500, color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer := bytes.Buffer{}

	err = imaging.Encode(&largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Create(ctx, "photos", "a.jpg", bytes.NewBuffer(largeImgBuffer.Bytes()))
	require.Nil(t, err)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{})
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err := imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds := img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{Method: "fit", Width: 300, Height: 300})
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, 300, imgBounds.X)
	require.Equal(t, 300, imgBounds.X)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{Method: "fit", Width: imgMaxWidth + 500, Height: imgMaxHeight + 500})
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)
}
