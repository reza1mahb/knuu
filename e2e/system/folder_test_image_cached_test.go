package system

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/celestiaorg/knuu/e2e"
	"github.com/celestiaorg/knuu/pkg/knuu"
)

func TestFolderCached(t *testing.T) {
	t.Parallel()

	// Setup
	executor, err := knuu.NewExecutor()
	if err != nil {
		t.Fatalf("Error creating executor: %v", err)
	}

	// Test logic
	const numberOfInstances = 10
	instances := make([]*knuu.Instance, numberOfInstances)

	for i := 0; i < numberOfInstances; i++ {
		instanceName := fmt.Sprintf("web%d", i+1)
		instances[i] = e2e.AssertCreateInstanceNginxWithVolumeOwner(t, instanceName)
	}

	var wgFolders sync.WaitGroup
	for _, instance := range instances {
		wgFolders.Add(1)
		go func(instance *knuu.Instance) {
			defer wgFolders.Done()
			// adding the folder after the Commit, it will help us to use a cached image.
			err := instance.AddFolder("resources/html", "/usr/share/nginx/html", "0:0")
			if err != nil {
				t.Errorf("Error adding file to '%v': %v", instance.Name(), err)
			}
		}(instance)
	}
	wgFolders.Wait()

	// Cleanup
	t.Cleanup(func() {
		err := e2e.AssertCleanupInstances(t, executor, instances)
		if err != nil {
			t.Fatalf("Error cleaning up: %v", err)
		}
	})

	// Test logic
	for _, instance := range instances {
		err = instance.StartAsync()
		if err != nil {
			t.Fatalf("Error waiting for instance to be running: %v", err)
		}
	}

	for _, instance := range instances {
		webIP, err := instance.GetIP()
		if err != nil {
			t.Fatalf("Error getting IP: %v", err)
		}

		err = instance.WaitInstanceIsRunning()
		if err != nil {
			t.Fatalf("Error waiting for instance to be running: %v", err)
		}

		wget, err := executor.ExecuteCommand("wget", "-q", "-O", "-", webIP)
		if err != nil {
			t.Fatalf("Error executing command: %v", err)
		}

		assert.Contains(t, wget, "Hello World!")
	}
}
