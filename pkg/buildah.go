package pkg

import (
	"context"
	"os"

	"go.podman.io/buildah"
	"go.podman.io/image/v5/types"
	storage "go.podman.io/storage"
	storage_types "go.podman.io/storage/types"
)

func GetBuildahStorage(homeDir string) (storage.Store, error) {
	// If you don't set storage and pass it into buildah.Pull
	// It'll fail with a segfault
	// Also, buildah (cli) uses ~/.local/share/containers/storage as the default storage location, so we need to set it to that so we can see images we pull via code in the cli
	store, err := storage.GetStore(storage_types.StoreOptions{RunRoot: "/run/user/1000", GraphRoot: homeDir + "/.local/share/containers/storage"})

	return store, err
}

func PullBuildahImage(store storage.Store, imageName string) (string, error) {
	// policy.json is related to https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md
	// move it to somewhere on the fs and read from it there specifically
	id, err := buildah.Pull(context.TODO(), "docker://"+imageName, buildah.PullOptions{ReportWriter: os.Stdout, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	return id, err
}

func PushBuildahImage(store storage.Store, imageID string, dest types.ImageReference) error {
	// Push the image id from the pull, earlier above
	_, _, err := buildah.Push(context.TODO(), imageID, dest, buildah.PushOptions{ReportWriter: os.Stdout, Store: store, SignaturePolicyPath: "/etc/containers/policy.json"})

	return err
}
