package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	runc "github.com/containerd/go-runc"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func BoolPointer(b bool) *bool {
	return &b
}

func RuncSetTerminal(homeDir string, image string, tag string, log *zap.SugaredLogger) error {
	// This entire function does the following
	// Opens the config.json file apart of the bundle that was created by umoci.Unpack
	// Unmarshals the json into a map[string]any
	// Accesses `process.terminal` and sets it to false
	// The default of this is 'true' - this will cause runc to fail when trying to start the container in detached mode (which we start it as)
	// Writes to the config.json file back to disk
	data, err := os.ReadFile(homeDir + "/code/buildah-go/bundle/" + image + "/" + tag + "/config.json")
	if err != nil {
		log.Error(err)
		return err
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		log.Error(err)
		return err
	}
	// Access the top level 'process' json property
	outer, ok := m["process"].(map[string]any)
	if !ok {
		log.Error("process property not found in config.json")
		return fmt.Errorf("process property not found in config.json")
	}
	log.Info("[runc] `process` property found in config.json")
	// Access the 'terminal' json property
	terminal, ok := outer["terminal"].(bool)
	if !ok {
		log.Error("terminal property not found in config.json")
		return fmt.Errorf("terminal property not found in config.json")
	}
	log.Infof("[runc] `terminal` found `config.json` with a value of: %v", terminal)
	// Set the 'terminal' pr
	terminal = false
	// Write back
	newBytes, err := json.Marshal(m)
	if err != nil {
		log.Error(err)
		return err
	}
	if err := os.WriteFile(homeDir+"/code/buildah-go/bundle/"+image+"/"+tag+"/config.json", newBytes, 0644); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func Runc(homeDir string, directory string, log *zap.SugaredLogger, image string, tag string) error {
	// Create a new UUID for the container ID
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}
	containerID := image + "-" + tag + "-" + newUUID.String()
	// Instantiate runc
	r := &runc.Runc{
		Rootless: BoolPointer(false),
	}
	// Since we run in detached mode, we also need to set `terminal = false` in `config.json`
	// There are no builtin cli or programmatic `go-runc` helpers for this
	// So the below function is called to manually set the `terminal` property in `config.json` to `false`
	if err := RuncSetTerminal(homeDir, image, tag, log); err != nil {
		return err
	}

	log.Info("[runc] Starting container via `runc` with bundle at " + homeDir + "/code/buildah-go/bundle/" + image + "/" + tag)
	opts := &runc.CreateOpts{
		// Start this in detached (background) mode
		Detach: true,
	}

	_, err2 := r.Run(context.Background(), containerID, homeDir+"/code/buildah-go/bundle/"+image+"/"+tag, opts)

	return err2
}
