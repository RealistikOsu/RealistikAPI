#!/usr/bin/make

build:
	docker build -t api:latest .

run-api:
	docker run \
		--env APP_COMPONENT=api \
		--network=host \
		--env-file=.env \
		-it api:latest