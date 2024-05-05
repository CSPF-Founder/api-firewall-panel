DOCKER_STACK = api-firewall

up:
	@echo "Starting containers..."
	docker compose -p ${DOCKER_STACK} up -d
	@echo "Containers started!"
# down: stops all containers
down:
	@echo "Stopping containers..."
	docker compose -p ${DOCKER_STACK} down
	@echo "Containers stopped!"
