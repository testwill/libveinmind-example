package walk

import (
	"io/fs"

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
		for i := 0; i < dockerImage.NumLayers(); i++ {

			l, err := dockerImage.OpenLayer(i)
			if err != nil {
				log.Error(err)
			}

			log.Info("Start Scan Layer: ", l.ID())
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
