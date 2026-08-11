# Schej.it API

API docs (available when the server is running): http://localhost:3002/swagger/index.html

## Debug

- Install PostgreSQL and create a database
- Set `DATABASE_URL` in `.env`
- Install `air`, a package that facilitates live reload for Go apps
  - `go install github.com/cosmtrek/air@latest`
- To run the server, simply run `air` in the root directory of the server

## Back up PostgreSQL

- Run `pg_dump "$DATABASE_URL" --format=custom --file=timeful.dump` to back up
- Run `pg_restore --clean --if-exists --dbname="$DATABASE_URL" timeful.dump` to restore
