package pkg

import (
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/oci/cas/dir"
	"github.com/opencontainers/umoci/oci/casext"
	"github.com/opencontainers/umoci/oci/layer"
	"go.uber.org/zap"
)

func UmociUnpack(homeDir string, log *zap.SugaredLogger, image string, tag string) error {
	e, err := dir.Open(homeDir + "/code/buildah-go/" + image)
	if err != nil {
		log.Error(err)
		return err
	}

	engineExt := casext.NewEngine(e)
	log.Info("[umoci] Attempting to unpack image to oci layout at " + homeDir + "/code/buildah-go/bundle/" + image + "/" + tag)
	err2 := umoci.Unpack(engineExt, tag, homeDir+"/code/buildah-go/bundle/"+image+"/"+tag, layer.UnpackOptions{})

	return err2
}
