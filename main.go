package main

import (
	"context"
	"os"

	buildah "go.podman.io/buildah"
	"go.podman.io/image/v5/transports/alltransports"
	storage "go.podman.io/storage"
	"go.podman.io/storage/pkg/reexec"
	storage_types "go.podman.io/storage/types"
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
	// Also, buildah (cli) uses ~/.local/share/containers/storage as the default storage location, so we need to set it to that so we can see images we pull via code in the cli
	store, err := storage.GetStore(storage_types.StoreOptions{RunRoot: "/run/user/1000", GraphRoot: os.Getenv("HOME") + "/.local/share/containers/storage"})
	if err != nil {
		log.Error(err)
		return
	}
	// policy.json is related to https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md
	// move it to somewhere on the fs and read from it there specifically
	id, err := buildah.Pull(context.TODO(), "docker://docker.io/redis:latest", buildah.PullOptions{ReportWriter: os.Stdout, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	if err != nil {
		log.Error(err)
	}

	log.Infof("Pulled image with image ID: %s", id)
	// TODO: figure out why os.Getenv("HOME") + "/code/buildah-go/bundle" doesn't work, but hardcoding it does
	log.Info("Attempting to push image to oci layout at /home/ajssalemo/code/buildah-go/bundle")

	dir := "oci:/home/ajssalemo/code/buildah-go/bundle"
	dest, err := alltransports.ParseImageName(dir)
	if err != nil {
		log.Error(err)
		return
	}

	_, _, err = buildah.Push(context.TODO(), "nginx:latest", dest, buildah.PushOptions{ReportWriter: os.Stdout, Store: store, BlobDirectory: os.Getenv("HOME") + "/code/buildah-go/blob", SignaturePolicyPath: "/etc/containers/policy.json"})
	if err != nil {
		log.Error(err)
	}

	log.Info("Successfully pushed image to oci layout at /home/ajssalemo/code/buildah-go/bundle")
}
