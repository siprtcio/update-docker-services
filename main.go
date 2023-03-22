package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	// Check if the new image name and stack name are provided as arguments
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run update-docker-services.go <new-image-name> <stack-name>")
	}
	newImageName := os.Args[1]
	stackName := os.Args[2]

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(), client.WithVersion("1.41"))
	if err != nil {
		log.Fatal(err)
	}

	// Set filters to search for services with a specific suffix
	filter := filters.NewArgs()
	filter.Add("label", "com.docker.stack.namespace="+stackName)

	// List all services that match the filters
	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{
		Filters: filter,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate with Amazon ECR
	authConfig := types.AuthConfig{
		Username: os.Getenv("AWS_ACCESS_KEY_ID"),
		Password: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	authConfigBytes, err := json.Marshal(authConfig)
	if err != nil {
		log.Fatal(err)
	}
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

	// Loop through all services and pull the new image to all swarm nodes
	for _, service := range services {
		// Get the current service configuration
		currentSpec := service.Spec

		// Pull the new image from Amazon ECR to all nodes in the swarm
		image := strings.Split(newImageName, ":")[0]
		tag := strings.Split(newImageName, ":")[1]
		imagePullResponse, err := cli.ImagePull(context.Background(), fmt.Sprintf("%s:%s", image, tag), types.ImagePullOptions{
			RegistryAuth: authConfigEncoded,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer imagePullResponse.Close()

		// Print a message indicating that the image was pulled to the swarm nodes
		fmt.Printf("Image %s pulled to all nodes in service %s\n", newImageName, service.Spec.Name)

		// Update the image in the service configuration
		newSpec := currentSpec
		newSpec.TaskTemplate.ContainerSpec.Image = newImageName

		// Update the service with the new configuration
		_, err = cli.ServiceUpdate(context.Background(), service.ID, service.Version, newSpec, types.ServiceUpdateOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// Print a message indicating that the service was updated
		fmt.Printf("Service %s updated to use image %s\n", service.Spec.Name, newImageName)
	}
}
