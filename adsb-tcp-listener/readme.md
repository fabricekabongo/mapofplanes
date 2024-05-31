This project is part of the Project Matrix
# ADSB TCP Listener

This is a service that connects to a remote TCP server that streams ADSB messages in the SBS1 format.
This service parses the message and post it to RabbitMQ as a JSON object. 

This server can handle very high RPM, working quite comfortabbly at 120k RPM and more.

# Set Up

## Environment Variables
 - ADSB_HOST
 - ADSB_PORT
 - RABBITMQ_URL
 - RABBITMQ_QUEUE

 ## Docker
 Run `docker buildx build -t IMAGE_NAME .`

