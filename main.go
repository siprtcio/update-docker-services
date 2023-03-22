package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	// Check if the new image name is provided as an argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run update-docker-services.go <new-image-name>")
	}
	newImageName := os.Args[1]

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	// Set filters to search for services with a specific suffix
	filter := filters.NewArgs()
	filter.Add("label", "com.docker.stack.namespace="+getStackNamespace())
	filter.Add("name", "*-suffix")

	// List all services that match the filters
	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{
		Filters: filter,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Loop through all services and update their image
	for _, service := range services {
		// Get the current service configuration
		currentSpec := service.Spec

		// Update the image in the service configuration
		newSpec := currentSpec
		newSpec.TaskTemplate.ContainerSpec.Image = newImageName

		// Update the service with the new configuration
		_, err := cli.ServiceUpdate(context.Background(), service.ID, service.Version, newSpec, types.ServiceUpdateOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// Print a message indicating that the service was updated
		fmt.Printf("Service %s updated to use image %s\n", service.Spec.Name, newImageName)
	}
}

// Helper function to get the Docker stack namespace
func getStackNamespace() string {
	namespace, found := os.LookupEnv("DOCKER_STACK_NAMESPACE")
	if !found {
		namespace = "default"
	}
	return namespace
}
