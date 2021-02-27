compose_build:
	docker-compose build

compose_start:
	docker-compose start

compose_down:
	docker-compose down

compose_restart:
	docker-compose down; \
	docker-compose build; \
	docker-compose start -d

bash_golang:
	docker exec -it golang bash

