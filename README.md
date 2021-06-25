# Idlemon Server

This is a game server for an online idle game. Player actions are received through HTTP requests and WebSockets are used to send realtime data to players. PostgreSQL is used to persist player data and Redis is used to cache temporary data.

### Dependencies

-   Go (1.16.x)
-   PostgreSQL (13.x)
-   Redis (6.x)

### Idlemon Client

The [Idlemon client](https://github.com/cdrpl/idlemon-client) can be used to interact with the server.

### Documentation

-   [API Documentation](https://documenter.getpostman.com/view/12308444/T1LLE7wE)

### Environment Variables

A .env file can be used to set environment variables. When the server is starting up it will try to read the .env file in the project root. Make a copy of [example.env](/example.env) and name it .env. Note that variables loaded from the .env file will never overwrite existing variables.

### CLI Flags

-   `-h` - Will display the list of CLI flags.
-   `-e [file]` - Use to specify the location of a .env file.
-   `-e nil` - This will stop the server from attempting to load a .env file.

### Docker

Run the following commands to build and run the server with Docker.

-   Build Image - `docker build -t idlemon .`
-   Run Container - `docker run -d --env-file .env --restart always --name idlemon -p 3000:3000 idlemon`

### Docker Compose

Docker compose is capable of setting up the server with a single command `docker compose up`. This will setup the server behind a NGINX reverse proxy, the server can be reached at localhost.

### Systemd

If you don't want to use Docker you can use systemd, an [example service file](/idlemon.service) is located in the project root.

### Authentication

Many API routes can only be accessed by authenticated users. When a user successfully logs in an API token is generated and returned in the HTTP response. To access a restricted route the user ID and API token must be included in the Authorization header separated by a colon. Example: "Authorization=ID:TOKEN".

### Database Tables

The server will construct the tables during startup. Just make sure a database exists with the same name as the env var `DB_NAME`. This feature can be disabled by setting the env var `CREATE_TABLES` to false.

### NGINX

NGINX can be used as a reverse proxy, access logger, rate limiter, and gzip compressor. An example [config](/nginx.conf) file is located in the root directory.

### Admin User Account

An admin account will be created during startup if the env var `CREATE_TABLES` is set to true. The admin account details can be set through the env vars `ADMIN_NAME`, `ADMIN_EMAIL`, and `ADMIN_PASS`.

## Development

### Setup Development Environment

1. Install Go.
2. Make sure you have access to a running instance of Postgres and Redis.
3. Make a copy of the [example.env](/example.env) file and name it ".env".
4. Enter the correct database credentials in the newly made .env file.
5. Make sure a database with the same name as the DB_NAME env var exists.
6. Run the server `go run .`

### WebSocket Server

Use the route `ws://localhost:3000/ws` when opening up a WebSocket connection with the server. Note that a valid authorization header is required to connect to the WebSocket server.

### Chat Messages

A player can initiate sending a chat message by making an HTTP request to an API endpoint. The server will store the chat message in the database then send the chat message to all connected WebSocket clients.
