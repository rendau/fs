package main

import (
	"archive/zip"
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
const testDirPath = "test_dir"
const imgMaxWidth = 1000
const imgMaxHeight = 1000

var (
	app = struct {
		lg   *zap.St
		core *core.St
	}{}
)

func cleanTestDir() {
	err := filepath.Walk(testDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if info.IsDir() {
			return os.RemoveAll(path)
		}

		return os.Remove(path)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	err = os.MkdirAll(testDirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

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
		testDirPath,
		imgMaxWidth,
		imgMaxHeight,
	)

	// Start tests
	code := m.Run()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	cleanTestDir()

	ctx := context.Background()

	fileContentRaw := []byte("test_data")

	fPath, err := app.core.Create(ctx, "photos", "data.txt", bytes.NewBuffer(fileContentRaw), false)
	require.Nil(t, err)

	fPathPrefix := "photos/" + time.Now().Format("2006/01/02") + "/"

	require.True(t, strings.HasPrefix(fPath, fPathPrefix))
	require.False(t, strings.Contains(strings.TrimPrefix(fPath, fPathPrefix), "/"))

	fName, fContent, err := app.core.Get(ctx, fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)
	require.Equal(t, "test_data", string(fContent))
	require.NotEmpty(t, fName)

	largeImg := imaging.New(imgMaxWidth+500, imgMaxHeight+500, color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer := new(bytes.Buffer)

	err = imaging.Encode(largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Create(ctx, "photos", "a.jpg", bytes.NewBuffer(largeImgBuffer.Bytes()), false)
	require.Nil(t, err)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err := imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds := img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{Method: "fit", Width: 300, Height: 300}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, 300, imgBounds.X)
	require.Equal(t, 300, imgBounds.X)

	fName, fContent, err = app.core.Get(ctx, fPath, &entities.ImgParsSt{Method: "fit", Width: imgMaxWidth + 500, Height: imgMaxHeight + 500}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)
}

func TestCreateZip(t *testing.T) {
	cleanTestDir()

	ctx := context.Background()

	zipFiles := [][2]string{
		{"index.html", "some html content"},
		{"abc/file.txt", "file content"},
		{"abc/qwe/x.txt", "x content"},
		{"todo.txt", "todo content"},
	}

	zipBuffer, err := createZipArchive(zipFiles)
	require.Nil(t, err)

	fPath, err := app.core.Create(ctx, "zip", "a.zip", zipBuffer, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fPath, "/"))

	for _, zp := range zipFiles {
		_, fContent, err := app.core.Get(ctx, fPath+zp[0], &entities.ImgParsSt{}, false)
		require.Nil(t, err)
		require.NotNil(t, fContent)
		require.Equal(t, zp[1], string(fContent))
	}

	fName, fContent, err := app.core.Get(ctx, fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.Equal(t, "index.html", fName)
	require.NotNil(t, fContent)
	require.Equal(t, "some html content", string(fContent))
}

func createZipArchive(items [][2]string) (*bytes.Buffer, error) {
	result := new(bytes.Buffer)

	zipWriter := zip.NewWriter(result)
	defer zipWriter.Close()

	for _, item := range items {
		f, err := zipWriter.Create(item[0])
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(item[1]))
		if err != nil {
			log.Fatal(err)
		}
	}

	return result, nil
}
