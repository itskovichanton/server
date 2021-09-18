package filestorage

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Работа с файловым хранилищем. Клиенты могут забирать из него файлы. При этом если
type IFileStorageService interface {

	// Возвращает файл из папки filestorage текущего проекта. Если передать lastModifiedTime>текущее время - вернет ошибку "Файл не нуждается в обновшении"
	// Например, мы уже скачали файл и нам не надо еще раз его скачивать (ReasonNoUpdateNeeded). Если файл не сущесвтует - возвращает ошибку (ReasonNotFound)
	GetFile(key string, lastModifiedTime *time.Time) (*FileInfo, error)
}

type FileInfo struct {
	FileInfo os.FileInfo
	FullPath string
}

const (
	ReasonNoUpdateNeeded = "REASON_NO_UPDATE_NEEDED"
	ReasonNotFound       = "REASON_NOT_FOUND"
)

type FileStorageError struct {
	errs.BaseError

	FilePath string
	Reason   string
}

type FileStorageService struct {
	IFileStorageService

	Config *core.Config
}

func (c *FileStorageService) GetFile(key string, lastModifiedTime *time.Time) (*FileInfo, error) {
	fileNamePrefix := utils.MD5(key) + "_"
	files, err := ioutil.ReadDir(c.Config.GetFileStorageDir())
	if err != nil {
		return nil, &FileStorageError{
			BaseError: *errs.NewBaseErrorFromCauseMsg(err, "Не удалось создать директорию "+c.Config.GetFileStorageDir()),
			Reason:    ReasonNotFound,
		}
	}

	var r os.FileInfo
	r = nil
	for _, f := range files {
		if strings.HasPrefix(f.Name(), fileNamePrefix) {
			r = f
			break
		}
	}

	if r != nil {
		fileName := filepath.Join(c.Config.GetFileStorageDir(), r.Name())
		dateStr := fileName[1+strings.Index(fileName, "_") : strings.Index(fileName, ".")]
		dateLong, err := strconv.ParseInt(dateStr, 10, 64)
		if err != nil {
			return nil, &FileStorageError{
				BaseError: *errs.NewBaseErrorFromCauseMsg(err, "Не удается найти файл"),
				Reason:    ReasonNotFound,
			}
		}
		if lastModifiedTime == nil || time.Unix(dateLong, 0).UTC().After(*lastModifiedTime) {
			return &FileInfo{
				FileInfo: r,
				FullPath: fileName,
			}, nil
		}

		return nil, &FileStorageError{
			BaseError: *errs.NewBaseErrorFromCauseMsg(err, "Файл не нуждается в обновлении"),
			FilePath:  fileName,
			Reason:    ReasonNoUpdateNeeded,
		}
	}

	return nil, &FileStorageError{
		BaseError: *errs.NewBaseErrorFromCauseMsg(err, "Файл не найден"),
		FilePath:  key,
		Reason:    ReasonNotFound,
	}
}
