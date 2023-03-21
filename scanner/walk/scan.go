package walk

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"strings"

	api "github.com/chaitin/libveinmind/go"
	"github.com/chaitin/libveinmind/go/docker"
	"github.com/chaitin/libveinmind/go/plugin/log"
)

func Scan(image api.Image) (err error) {

	refs, err := image.RepoRefs()
	var imageRef string
	if err == nil && len(refs) > 0 {
		imageRef = refs[0]
	} else {
		imageRef = image.ID()
	}
	log.Info("Scan Image: ", imageRef)

	// 判断是否可以获取 Layer
	switch v := image.(type) {
	case *docker.Image:
		dockerImage := v
		var (
			parentChanId string
			hostPath     string
		)

		for i := 0; i < dockerImage.NumLayers(); i++ {
			var l api.Layer
			l, err = dockerImage.OpenLayer(i)
			if err != nil {
				log.Error(err)
			}
			if i == 0 {
				hostPath, parentChanId, err = transferImagePath(l.ID(), "")

			} else {
				hostPath, parentChanId, err = transferImagePath(l.ID(), parentChanId)
			}
			if err != nil {
				log.Error(err)
				continue
			}
			log.Info("Start Scan Layer: ", l.ID(), ", host path :", hostPath, ", parentChanId :", parentChanId)
			l.Walk("/", func(path string, info fs.FileInfo, err error) error {
				defer func() {
					if err := recover(); err != nil {
						log.Error(err)
					}
				}()

				// 处理错误
				if err != nil {
					log.Debug(err)
					return nil
				}
				log.Debug("path :", path)
				f, err := l.Open(path)
				if err != nil {
					log.Debug(err)
					return nil
				}

				defer func() {
					f.Close()
				}()
				return nil
			})
		}
	}

	return nil
}

const (
	layerDbPath  = "/var/lib/docker/image/overlay2/layerdb/sha256/"
	cacheIdPath  = "/cache-id"
	overlay2Path = "/var/lib/docker/overlay2/"
	diffPath     = "/diff"
)

func transferImagePath(diffId, parentChanId string) (path, chanId string, err error) {
	var (
		id      []byte
		cacheId string
	)
	diffId = strings.TrimPrefix(diffId, "sha256:")
	parentChanId = strings.TrimPrefix(parentChanId, "sha256:")

	if parentChanId == "" {
		chanPath := layerDbPath + diffId + cacheIdPath
		if _, err = os.Stat(chanPath); errors.Is(err, os.ErrNotExist) {
			log.Error(err)
			return
		} else {
			id, err = os.ReadFile(chanPath)
			if err != nil {
				log.Error(err)
				return
			}
			cacheId = string(id)
			path = overlay2Path + string(cacheId) + diffPath
			if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
				log.Error(err)
				return
			}
			chanId = diffId //第一层用原始diffId
			return
		}
	} else {
		hash := sha256.Sum256([]byte("sha256:" + parentChanId + " " + "sha256:" + diffId))
		chanId = hex.EncodeToString(hash[:])
		log.Info("chanId :", chanId)
		chanPath := layerDbPath + chanId + cacheIdPath
		if _, err = os.Stat(chanPath); errors.Is(err, os.ErrNotExist) {
			log.Error(err)
			return
		} else {
			id, err = os.ReadFile(chanPath)
			if err != nil {
				log.Error(err)
				return
			}
			cacheId = string(id)
			path = overlay2Path + string(cacheId) + diffPath
			if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
				log.Error(err)
				return
			}
			return
		}
	}

	return
}
