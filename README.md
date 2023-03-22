# update-docker-services
golang script to update all docker services with same suffix with new image. 

Here's how to use the script:

1. Save the script to a file named update-docker-services.go.
2. Make the script executable with chmod +x update-docker-services.go.
3. Run the script with go run update-docker-services.go <new-image-name>, where <new-image-name> is the name of the new Docker image to use.
4. The script will search for all Docker services in the current stack with a label com.docker.stack.namespace that matches the stack namespace, and a name that ends with -suffix. It will then update each service to use the new image name provided as an argument. The script will output a message for each service that was updated.
