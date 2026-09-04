package pkg

import (
	"context"
	"os"

	"go.podman.io/buildah"
	"go.podman.io/image/v5/types"
	storage "go.podman.io/storage"
	storage_types "go.podman.io/storage/types"
	"go.uber.org/zap"
)

func GetBuildahStorage(homeDir string) (storage.Store, error) {
	// If you don't set storage and pass it into buildah.Pull
	// It'll fail with a segfault
	// Also, buildah (cli) uses ~/.local/share/containers/storage as the default storage location, so we need to set it to that so we can see images we pull via code in the cli
	store, err := storage.GetStore(storage_types.StoreOptions{RunRoot: "/run/user/1000", GraphRoot: homeDir + "/.local/share/containers/storage"})

	return store, err
}

func PullBuildahImage(store storage.Store, registryUrl string, imageName string, tag string, log *zap.SugaredLogger) (string, error) {
	// policy.json is related to https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md
	// move it to somewhere on the fs and read from it there specifically
	fullyQualifiedImageName := registryUrl + "/" + imageName + ":" + tag
	log.Info("[buildah] Pulling image " + fullyQualifiedImageName + " with `buildah pull`")
	id, err := buildah.Pull(context.TODO(), "docker://"+fullyQualifiedImageName, buildah.PullOptions{ReportWriter: os.Stdout, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	return id, err
}

func PushBuildahImage(store storage.Store, imageID string, dest types.ImageReference, log *zap.SugaredLogger, directory string) error {
	// Push the image id from the pull, earlier above
	log.Info("[buildah] Attempting to push image to oci layout at " + directory)
	_, _, err := buildah.Push(context.TODO(), imageID, dest, buildah.PushOptions{ReportWriter: os.Stdout, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	return err
}
