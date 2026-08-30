package main

import (
	"os"
	"strings"

	"go.podman.io/image/v5/transports/alltransports"
	"go.podman.io/storage/pkg/reexec"
	zap "go.uber.org/zap"

	pkg "buildah-go/pkg"
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
	// Get the users home directory
	// Since this is going to be ran as root typically (e.x `sudo $(which go) run . or `sudo ./somebinary`)
	// then it'll be under /root
	// Otherwise, its /home/<user>/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
		return
	}

	// If you don't set storage and pass it into buildah.Pull
	// It'll fail with a segfault
	// Also, buildah (cli) uses ~/.local/share/containers/storage as the default storage location, so we need to set it to that so we can see images we pull via code in the cli
	store, err := pkg.GetBuildahStorage(homeDir)
	if err != nil {
		log.Error(err)
		return
	}
	// policy.json is related to https://github.com/containers/image/blob/main/docs/containers-policy.json.5.md
	// move it to somewhere on the fs and read from it there specifically
	id, err := pkg.PullBuildahImage(store, "redis:latest", log)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("[buildah] Pulled image with image ID: %s", id)
	// The logic of `dir` and how this maps to being able to have runc find/create containers from it is is the following:
	// e.g `oci:/path/to/your/dir:image:tag`
	// 1. this must be prefixed with `oci:` - this will create files/folders in the directory specific in a OCI layout
	// 2. the path specified, for semantic reasons, should match the image name (but it doesnt have to): https://github.com/opencontainers/umoci/blob/main/doc/man/umoci-unpack.1.md
	// ex. /home/images/redis
	// 3. if tag is omitted, umoci looks up last. you can push multiple tags to the same directory
	// ex. when umoci looks it up, it would be like this: `umoci unpack --image /home/images/redis:1.2.3 /path/to/unpack/dir`
	// --------------------------------------------------- //
	// TODO: remove image hardcoding name
	directory := "oci:" + homeDir + "/code/buildah-go/redis:latest"
	dest, err := alltransports.ParseImageName(directory)
	if err != nil {
		log.Error(err)
		return
	}
	// Push the image id from the pull, earlier above
	buildahErr := pkg.PushBuildahImage(store, id, dest, log, homeDir)
	if buildahErr != nil {
		log.Error(buildahErr)
		return
	}
	log.Info("[buildah] Successfully pushed image to oci layout at " + homeDir + "/code/buildah-go/redis")
	// Switch to using umoci to create an OCI layout for runc to use
	// Use umoci to unpack the image to a directory that runc can use to create a container from
	umociErr := pkg.UmociUnpack(homeDir, log)
	if umociErr != nil {
		if strings.Contains(umociErr.Error(), "config.json already exists") {
			log.Warn("[umoci] config.json already exists, skipping unpack")
			// no-op
		} else {
			log.Error(umociErr)
			return
		}
	}
	log.Info("[umoci] Successfully unpacked image to oci layout at " + homeDir + "/code/buildah-go/bundle/redis")
	// Point runc to the bundle and create + start a container from it
	runcErr := pkg.Runc(homeDir, log)
	if runcErr != nil {
		log.Error(runcErr)
		return
	}
	log.Info("[runc] Successfully started container via `runc` with bundle at " + homeDir + "/code/buildah-go/bundle/redis")
}
