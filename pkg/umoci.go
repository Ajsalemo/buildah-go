package pkg

import (
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/oci/cas/dir"
	"github.com/opencontainers/umoci/oci/casext"
	"github.com/opencontainers/umoci/oci/layer"
	"go.uber.org/zap"
)

func UmociUnpack(homeDir string, log *zap.SugaredLogger) error {
	e, err := dir.Open(homeDir + "/code/buildah-go/redis")
	if err != nil {
		log.Error(err)
		return err
	}

	engineExt := casext.NewEngine(e)
	//  TODO: unpack this in the format of homeDir+"/code/buildah-go/bundle/{image}/{tag}
	log.Info("[umoci] Attempting to unpack image to oci layout at " + homeDir + "/code/buildah-go/redis")
	err2 := umoci.Unpack(engineExt, "latest", homeDir+"/code/buildah-go/bundle/redis", layer.UnpackOptions{})

	return err2
}
