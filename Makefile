# 加载 .env 文件
include .env
export $(shell sed 's/=.*//' .env)

# migrate
install_migrate:
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate_up:
	migrate -path="./database/migrate" -database=${DATABASE_URL} up

migrate_down:
	migrate -path="./database/migrate" -database=${DATABASE_URL} down

# sqlc
# install_sqlc:
# 	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# postgres
# install_postgres:
# 	docker run --name postgres_urls \
# 	-e POSTGRES_USER=root \
# 	-e POSTGRES_PASSWORD=password \
# 	-e POSTGRES_DB=urldb \
# 	-p 5432:5432 \
# 	-d postgres

# redis
launch_redis:
	docker run --name