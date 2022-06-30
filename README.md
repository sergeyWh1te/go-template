# go-template
go-skeleton-lido

 ## How to use the template
 1. Clone repository
 2. cd root repository
 3. make tools
 4. make vendor
 5. docker-composer up -d
 6. make migrate
 7. make build
 8. Run service ./bin/service

## How to create migrations?
 ./bin/migrate create -ext=sql -dir=db/migrations <your table name>

## How to make migrations?
1. make migrate from terminal or 
```
    bin/migrate -database ${POSTGRESQL_URL} -path db/migrations up
```

## Where I have to start add my handler ?
Here: /internal/pkg/your_package_name/. Just see an example with [User package](./internal/pkg/users)

## Docs and rules
1. [App structure layout](./docs/structure.md)
2. [Code style](./docs/code_style.md)