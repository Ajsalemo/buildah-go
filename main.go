package main

import (
	"context"
	"os"

	buildah "go.podman.io/buildah"
	storage "go.podman.io/storage"
	"go.podman.io/storage/pkg/reexec"
	types "go.podman.io/storage/types"
	zap "go.uber.org/zap"
)

func main() {
	// Important
	// See: https://github.com/containers/storage/blob/main/pkg/reexec/reexec.go#L48
	// Otherwise we panic when trying to unpack the image to fs storage
	if reexec.Init() {
		return
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()
	// If you don't set storage and pass it into buildah.Pull
	// It'll fail with a segfault
	store, err := storage.GetStore(types.StoreOptions{RunRoot: "/run/user/1000", GraphRoot: "/var/lib/containers/storage"})
	if err != nil {
		log.Error(err)
		return
	}
	// policy.json is related to https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md
	// move it to somewhere on the fs and read from it there specifically
	id, err := buildah.Pull(context.TODO(), "docker://docker.io/nginx:latest", buildah.PullOptions{ReportWriter: os.Stderr, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	if err != nil {
		log.Error(err)
	}

	log.Info(id)
}
