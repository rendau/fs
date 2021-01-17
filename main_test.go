package main

import (
	"archive/zip"
	"bytes"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/disintegration/imaging"
	cleanerMock "github.com/rendau/fs/internal/adapters/cleaner/mock"
	"github.com/rendau/fs/internal/adapters/logger/zap"
	"github.com/rendau/fs/internal/cns"
	"github.com/rendau/fs/internal/domain/core"
	"github.com/rendau/fs/internal/domain/entities"
	"github.com/rendau/fs/internal/domain/errs"
	"github.com/rendau/fs/internal/domain/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const confPath = "test_conf.yml"
const testDirPath = "test_dir"
const imgMaxWidth = 1000
const imgMaxHeight = 1000

var (
	app = struct {
		lg      *zap.St
		cleaner *cleanerMock.St
		core    *core.St
	}{}
)

func cleanTestDir() {
	err := filepath.Walk(testDirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if p == testDirPath {
			return nil
		}

		// app.lg.Infow("cleanTestDir walk", "path", p, "name", info.Name(), "mt", info.ModTime())

		if info.IsDir() {
			err = os.RemoveAll(p)
			if err != nil {
				return err
			}

			return filepath.SkipDir
		}

		return os.Remove(p)
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

	app.cleaner = cleanerMock.New()

	app.core = core.New(
		app.lg,
		testDirPath,
		imgMaxWidth,
		imgMaxHeight,
		"",
		0,
		[]string{},
		app.cleaner,
		true,
	)

	// Start tests
	code := m.Run()

	cleanTestDir()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	cleanTestDir()

	_, err := app.core.Create("asd/"+cns.ZipDirNamePrefix+"_asd", "a.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	_, err = app.core.Create(cns.ZipDirNamePrefix+"_asd/asd", "a.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	fPath, err := app.core.Create("photos", "data.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.Nil(t, err)

	fPathPrefix := "photos/" + time.Now().Format("2006/01/02") + "/"

	require.True(t, strings.HasPrefix(fPath, fPathPrefix))
	require.False(t, strings.Contains(strings.TrimPrefix(fPath, fPathPrefix), "/"))

	fName, _, fContent, err := app.core.Get(fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)
	require.Equal(t, "test_data", string(fContent))
	require.NotEmpty(t, fName)

	largeImg := imaging.New(imgMaxWidth+10, imgMaxHeight+10, color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer := new(bytes.Buffer)

	err = imaging.Encode(largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Create("photos", "a.jpg", largeImgBuffer, true, false)
	require.Nil(t, err)

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err := imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds := img.Bounds().Max
	require.Equal(t, imgMaxWidth+10, imgBounds.X)
	require.Equal(t, imgMaxHeight+10, imgBounds.X)

	largeImg = imaging.New(imgMaxWidth+10, imgMaxHeight+10, color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer = new(bytes.Buffer)

	err = imaging.Encode(largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Create("photos", "a.jpg", largeImgBuffer, false, false)
	require.Nil(t, err)

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{Method: "fit", Width: imgMaxWidth - 10, Height: imgMaxHeight - 10}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth-10, imgBounds.X)
	require.Equal(t, imgMaxHeight-10, imgBounds.X)

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{Method: "fit", Width: imgMaxWidth + 10, Height: imgMaxHeight + 10}, false)
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

	zipContentIsSame := func(a, b [][2]string) {
		require.Equal(t, len(a), len(b))

		for _, ai := range a {
			found := false

			for _, bi := range b {
				if bi[0] == ai[0] {
					found = true
					require.Equal(t, ai[1], bi[1])
					break
				}
			}

			require.True(t, found)
		}

		for _, bi := range b {
			found := false

			for _, ai := range a {
				if ai[0] == bi[0] {
					found = true
					require.Equal(t, bi[1], ai[1])
					break
				}
			}

			require.True(t, found)
		}
	}

	srcZipFiles := [][2]string{
		{"index.html", "some html content"},
		{"abc/file.txt", "file content"},
		{"abc/qwe/x.txt", "x content"},
		{"todo.txt", "todo content"},
	}

	zipBuffer, err := createZipArchive(srcZipFiles)
	require.Nil(t, err)

	_, err = app.core.Create("zip/"+cns.ZipDirNamePrefix+"_asd", "a.zip", zipBuffer, false, true)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	_, err = app.core.Create(cns.ZipDirNamePrefix+"_asd/zip", "a.zip", zipBuffer, false, true)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	fPath, err := app.core.Create("zip", "a.zip", zipBuffer, false, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fPath, "/"))

	for _, zp := range srcZipFiles {
		_, _, fContent, err := app.core.Get(fPath+zp[0], &entities.ImgParsSt{}, false)
		require.Nil(t, err)
		require.NotNil(t, fContent)
		require.Equal(t, zp[1], string(fContent))
	}

	fName, _, fContent, err := app.core.Get(fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.Equal(t, "index.html", fName)
	require.NotNil(t, fContent)
	require.Equal(t, "some html content", string(fContent))

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{}, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fName, ".zip"))
	require.NotNil(t, fContent)

	resultZipFiles, err := extractZipArchive(fContent)
	require.Nil(t, err)
	zipContentIsSame(srcZipFiles, resultZipFiles)

	srcZipFiles = [][2]string{
		{"root/index.html", "some html content"},
		{"root/abc/file.txt", "file content"},
		{"root/abc/qwe/x.txt", "x content"},
		{"root/todo.txt", "todo content"},
	}

	zipBuffer, err = createZipArchive(srcZipFiles)
	require.Nil(t, err)

	fPath, err = app.core.Create("zip", "a.zip", zipBuffer, false, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fPath, "/"))

	for _, zp := range srcZipFiles {
		_, _, fContent, err := app.core.Get(fPath+strings.TrimLeft(zp[0], "root/"), &entities.ImgParsSt{}, false)
		require.Nil(t, err)
		require.NotNil(t, fContent)
		require.Equal(t, zp[1], string(fContent))
	}

	fName, _, fContent, err = app.core.Get(fPath, &entities.ImgParsSt{}, false)
	require.Nil(t, err)
	require.Equal(t, "index.html", fName)
	require.NotNil(t, fContent)
	require.Equal(t, "some html content", string(fContent))
}

func TestClean(t *testing.T) {
	cleanTestDir()

	dirStructure := [][2]string{
		{"dir1", ""},
		{"dir2/file1.txt", "file1 content"},
		{"dir2/dir3/file2.txt", "file2 content"},
		{"file3.txt", "file3 content"},
		{"dir5/" + cns.ZipDirNamePrefix + "q/a/js.js", "content"},
		{"dir5/" + cns.ZipDirNamePrefix + "q/index.html", "content"},
		{"dir5/" + cns.ZipDirNamePrefix + "q/css.css", "content"},
		{"dir6/q" + cns.ZipDirNamePrefix + "/index.html", "content"},
	}

	err := makeDirStructure(testDirPath, dirStructure)
	require.Nil(t, err)

	compareDirStructure(t, testDirPath, dirStructure)

	checkedFiles := make([]string, 0)

	app.cleaner.SetHandler(func(pathList []string) []string {
		checkedFiles = append(checkedFiles, pathList...)
		return []string{}
	})

	app.core.Clean(0)

	dirStructure = dirStructure[1:]

	compareStringSlices(t, checkedFiles, []string{
		"dir2/file1.txt",
		"dir2/dir3/file2.txt",
		"file3.txt",
		"dir5/" + cns.ZipDirNamePrefix + "q/",
		"dir6/q" + cns.ZipDirNamePrefix + "/index.html",
	})

	compareDirStructure(t, testDirPath, dirStructure)

	app.cleaner.SetHandler(func(pathList []string) []string {
		return []string{
			dirStructure[1][0],
		}
	})

	app.core.Clean(0)

	dirStructure = append(dirStructure[:1], dirStructure[2:]...)

	compareDirStructure(t, testDirPath, dirStructure)

	app.cleaner.SetHandler(func(pathList []string) []string {
		return pathList
	})

	app.core.Clean(0)

	dirStructure = [][2]string{}

	compareDirStructure(t, testDirPath, dirStructure)

	err = makeDirStructure(testDirPath, [][2]string{
		{"dir1", ""},
		{"dir2", ""},
		{"dir2/dir3", ""},
		{"dir2/dir3/dir4", ""},
		{"dir2/dir5/dir6", ""},
		{"dir2/dir5/file1.txt", "asd"},
	})
	require.Nil(t, err)

	app.cleaner.SetHandler(func(pathList []string) []string { return []string{} })

	app.core.Clean(0)

	compareDirStructure(t, testDirPath, [][2]string{
		{"dir2/dir5/file1.txt", "asd"},
	})
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

func extractZipArchive(data []byte) ([][2]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	result := make([][2]string, 0)

	fileHandler := func(f *zip.File) error {
		if f.FileInfo().IsDir() {
			return nil
		}

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		srcFileDataRaw, err := ioutil.ReadAll(srcFile)
		if err != nil {
			return err
		}

		result = append(result, [2]string{f.Name, string(srcFileDataRaw)})

		return nil
	}

	for _, f := range reader.File {
		err = fileHandler(f)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func compareDirStructure(t *testing.T, dirPath string, items [][2]string) {
	diskItems := make([][2]string, 0)

	err := filepath.Walk(dirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		if p == dirPath {
			return nil
		}

		relP, err := filepath.Rel(dirPath, p)
		if err != nil {
			return err
		}

		relUrlP := util.ToUrlPath(relP)

		if info.IsDir() {
			diskItems = append(diskItems, [2]string{relUrlP, ""})
		} else {
			fileDataRaw, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}

			diskItems = append(diskItems, [2]string{relUrlP, string(fileDataRaw)})
		}

		return nil
	})
	require.Nil(t, err)

	for _, dItem := range diskItems {
		found := false
		dItemIsDir := path.Ext(dItem[0]) == ""

		for _, item := range items {
			if dItemIsDir {
				if strings.Contains(item[0], dItem[0]) {
					found = true
					break
				}
			} else {
				if item[0] == dItem[0] {
					require.Equal(t, item[1], dItem[1])
					found = true
					break
				}
			}
		}

		require.True(t, found, "Item not found %s", dItem[0])
	}

	for _, item := range items {
		found := false

		for _, dItem := range diskItems {
			if item[0] == dItem[0] {
				require.Equal(t, item[1], dItem[1])
				found = true
				break
			}
		}

		require.True(t, found, "Item not found %s", item[0])
	}
}

func makeDirStructure(parentDirPath string, items [][2]string) error {
	var err error

	for _, item := range items {
		fsPath := util.ToFsPath(item[0])

		if path.Ext(item[0]) == "" { // dir
			err = os.MkdirAll(filepath.Join(parentDirPath, fsPath), os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Join(parentDirPath, filepath.Dir(fsPath)), os.ModePerm)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filepath.Join(parentDirPath, fsPath), []byte(item[1]), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func compareStringSlices(t *testing.T, a, b []string) {
	require.Equal(t, len(a), len(b))

	for _, aI := range a {
		found := false

		for _, bI := range b {
			if bI == aI {
				found = true
				break
			}
		}

		require.True(t, found, "String not found %q", aI)
	}

	for _, bI := range b {
		found := false

		for _, aI := range a {
			if aI == bI {
				found = true
				break
			}
		}

		require.True(t, found, "String not found %q", bI)
	}
}
