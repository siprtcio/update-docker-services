package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
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

	// Get the list of nodes in the Docker Swarm cluster
	nodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		panic(err)
	}

	// Pull the Docker image on each node in the swarm
	for _, node := range nodes {
		// Check if the node is a manager node
		if node.Spec.Role == swarm.NodeRoleManager {
			// Pull the Docker image on the node
			reader, err := cli.ImagePull(context.Background(), newImageName, types.ImagePullOptions{
				RegistryAuth: authConfigEncoded,
			})
			if err != nil {
				fmt.Printf("Failed to pull image on node %s: %s\n", node.ID, err.Error())
			} else {
				io.Copy(os.Stdout, reader)
				fmt.Printf("Successfully pulled image on node %s\n", node.ID)
			}
		}
	}

	// Loop through all services and pull the new image to all swarm nodes
	for _, service := range services {
		// Get the current service configuration
		currentSpec := service.Spec

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
