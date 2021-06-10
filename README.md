# Idlemon Server

This is a game server for an online idle game. Most player actions are received through HTTP requests but WebSockets are used for realtime needs such as player chat. Player data is saved using MariaDB and Redis is used as a temporary cache.

### Dependencies

-   Go (1.16.x)
-   MariaDB (10.x)
-   Redis (6.x)

### Documentation

-   [API Documentation](https://documenter.getpostman.com/view/12308444/T1LLE7wE)

### Environment Variables

A .env file can be used to set environment variables. When the server is starting up it will try to read the .env file in the project root. Make a copy of [example.env](/example.env) and name it .env. Note that variables loaded from the .env file will never overwrite existing variables.

### CLI Flags

-   `-h` - Will display the list of CLI flags.
-   `-e [file]` - Use to specify the location of a .env file.
-   `-e nil` - This will stop the server from attempting to load a .env file.
-   `-d` - this will cause all tables to be dropped then recreated during startup. A QoL feature for development and will be removed in later versions.

### Docker

Run the following commands to build and run the server with Docker.

-   Build Image - `docker build -t idlemon .`
-   Run Container - `docker run -d --env-file .env --restart always --name idlemon -p 3000:3000 idlemon`

### Docker Compose

Docker compose is capable of setting up the server with a single command `docker compose up`. This will setup the server behind a NGINX reverse proxy, the server can be reached at localhost. phpMyAdmin will also be setup and can be accessed at localhost:8080. To login, use the root user and the password set with the `DB_PASS` environment variable.

### Authentication

Many API routes can only be accessed by authenticated users. When a user successfully logs in an API token is generated and returned in the HTTP response. To access a restricted route the user ID and API token must be included in the Authorization header separated by a colon. Example: "Authorization=ID:TOKEN".

### Database Tables

The server will construct the tables during startup. Just make sure a database exists with the same name as the env var `DB_NAME`.

### NGINX

NGINX can be used as a reverse proxy, access logger, and gzip compressor. An example [config](/nginx.conf) file is located in the root directory.

### Admin User Account

The server will create an admin account when starting up. The password is set using the `ADMIN_PASS` environment variable. This is a special account used by the server. Since units cannot exist without an owner and the server needs units for the campaign system, this account is used as the owner for all server owned units.

## Development

### Setup Development Environment

1. Install Go.
2. Make sure you have access to a running instance of MariaDB and Redis.
3. Make a copy of the [example.env](/example.env) file and name it ".env".
4. Enter the correct database credentials in the newly made .env file.
5. Make sure a database with the same name as the DB_NAME env var exists.
6. Run the server `go run .`

### WebSocket Server

Use the route `ws://localhost:3000/ws` when opening up a WebSocket connection with the server.
