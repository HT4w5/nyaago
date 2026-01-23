package fsr

import (
	"os"
	"strings"

	"github.com/HT4w5/nyaago/internal/logging"
)

// Request size of a path. If not cached, request the file, but don't block
func (fsr *FileSendRatio) getPathSize(path string) (int64, bool) {
	// Lookup cache
	rec, err := fsr.getFileSizeRecord(path)
	if err != nil {
		if err == ErrRecordNotFound {
			fsr.logger.Info("file size cache miss", "path", path)
			go fsr.readPathSizeFromDisk(path)
		} else {
			fsr.logger.Error("failed to get fileSizeRecord", logging.SlogKeyError, err)
		}
		return 0, false
	} else {
		return rec.Size, true
	}
}

func (fsr *FileSendRatio) readPathSizeFromDisk(path string) {
	var newPath string
	found := false
	for _, v := range fsr.cfg.PathMap {
		if !strings.HasPrefix(path, v.UrlPrefix) {
			continue
		}
		newPath = v.DirPrefix + strings.TrimPrefix(path, v.UrlPrefix)
		found = true
		break
	}

	if !found {
		fsr.logger.Warn("dropping path without prefix in path map", "path", path)
		return
	}

	// Read size from disk
	info, err := os.Stat(newPath)
	if err != nil {
		fsr.logger.Error("failed to get file size", logging.SlogKeyError, err, "path", path)
		return
	}

	err = fsr.putFileSizeRecord(fileSizeRecord{
		Path: path,
		Size: info.Size(),
	})

	if err != nil {
		fsr.logger.Error("failed to put file size record", logging.SlogKeyError, err)
		return
	}
}
