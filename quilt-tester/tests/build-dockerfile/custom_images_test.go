package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/connection/credentials"
	"github.com/quilt/quilt/db"
)

func TestCustomImages(t *testing.T) {
	clnt, err := client.New(api.DefaultSocket, credentials.Insecure{})
	if err != nil {
		t.Fatalf("couldn't get quiltctl client: %s", err)
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err)
	}

	machines, err := clnt.QueryMachines()
	if err != nil {
		t.Fatalf("couldn't query machines: %s", err)
	}

	// The images deployed for the given Dockerfile.
	dockerfileToImages := make(map[string][]string)

	// The number of containers deployed for each Dockerfile.
	dockerfileCount := make(map[string]int)

	for _, c := range containers {
		if !strings.Contains(c.Image, "test-custom-image") {
			continue
		}

		dockerfileID, imageID, err := getContainerInfo(c.StitchID)
		if err != nil {
			t.Fatalf("couldn't get container info for %s: %s",
				c.StitchID, err)
		}

		dockerfileToImages[dockerfileID] = append(
			dockerfileToImages[dockerfileID], imageID)
		dockerfileCount[dockerfileID]++
	}

	fmt.Println("Dockerfile to image mappings:", dockerfileToImages)
	fmt.Println("Dockerfile counts:", dockerfileCount)

	reuseErr := checkReuseImage(dockerfileToImages)
	if reuseErr != nil {
		t.Error(reuseErr)
	}

	countErr := checkImageCounts(machines, dockerfileCount)
	if countErr != nil {
		t.Error(countErr)
	}
}

func checkReuseImage(dockerfileToImages map[string][]string) error {
	for dk, images := range dockerfileToImages {
		for _, otherImg := range images {
			if otherImg != images[0] {
				return fmt.Errorf("images for DockerfileID %s not "+
					"reused: %v", dk, images)
			}
		}
	}
	return nil
}

func checkImageCounts(machines []db.Machine, dockerfileCounts map[string]int) error {
	nWorker := 0
	for _, m := range machines {
		if m.Role == db.Worker {
			nWorker++
		}
	}

	for i := 0; i < nWorker; i++ {
		if actual := dockerfileCounts[strconv.Itoa(i)]; actual != 2 {
			return fmt.Errorf("DockerfileID %d had %d containers, "+
				"expected %d", i, actual, 2)
		}
	}

	return nil
}

func getContainerInfo(stitchID string) (string, string, error) {
	dockerfileID, err := exec.Command(
		"quilt", "ssh", stitchID, "cat /dockerfile-id").CombinedOutput()
	if err != nil {
		return "", "", err
	}
	imageID, err := exec.Command(
		"quilt", "ssh", stitchID, "cat /image-id").CombinedOutput()
	return strings.TrimSpace(string(dockerfileID)),
		strings.TrimSpace(string(imageID)),
		err
}
