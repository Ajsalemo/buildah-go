package pkg

import (
	"context"

	runc "github.com/containerd/go-runc"
	"go.uber.org/zap"
)

func BoolPointer(b bool) *bool {
	return &b
}

func Runc(homeDir string, log *zap.SugaredLogger) error {

	r := &runc.Runc{
		Rootless: BoolPointer(false),
	}

	log.Info("[runc] Starting container via `runc` with bundle at " + homeDir + "/code/buildah-go/bundle/redis")
	opts := &runc.CreateOpts{
		// Start this in detached (background) mode
		// Or else this stays in the foreground and makes getting out of the terminal difficult
		Detach: true,
	}
	_, err := r.Run(context.Background(), "redis-0", homeDir+"/code/buildah-go/bundle/redis", opts)

	return err
}
