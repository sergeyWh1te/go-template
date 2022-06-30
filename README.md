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

## Docs and rules
1. [App_structure layout](./docs/structure.md) of project structure layout
2. [Code style](./docs/code_style.md)