# update-docker-services
golang script to update all docker services with same suffix with new image. 

1. Save the modified script to a file named update-docker-services.go.
2. Make the script executable with chmod +x update-docker-services.go.
3. Run the script with go run update-docker-services.go <new-image-name> <service-suffix>, where <new-image-name> is the name of the new Docker image to use and <service-suffix> is the suffix that identifies the services to
