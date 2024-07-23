.PHONY: run-postgres, stop-postgres, rm-postgres

run-postgres:
	docker run \
	--name postgres \
	-p 5432:5432 \
	-e POSTGRES_PASSWORD=root \
	-e POSTGRES_USER=root \
	-d postgres:alpine3.20

stop-postgres:
	docker stop postgres

start-postgres:
	docker start postgres

rm-postgres:
	make stop-postgres \
	&& docker rm postgres
