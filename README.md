
# Auth

**Wiki** : https://app.theneo.io/szeeta/auth

----------

# How do I run this locally ?

## Prerequisites

- Docker
  - With docker-compose
  - You should be able to run docker as a deamon

- Golang

## Steps

1. Copy the contents of all the .env.example files to .env files in the relevant locations and then modify them as needed

```
cp ./backend/.env.example ./backend/.env
cp ./services/.env.example ./services/.env

# Modify the content of the env files as needed
# Eg :- Adding your own resend API key
```

2. Run the docker services

```
cd ./services
# Make sure your docker deamon is up and running
docker compose up
```

3. Then go to the backend directory and run the server with golang

```
cd ./backend
go run cmd/main.go
```
 
----------

Please open a GitHub issue if you have any suggestions or comments
