# Tinydates
Muzz technical assessment by Ebuka Agbanyim

## How to run
This application is built and deployed user docker, a [`docker compose`](./docker-compose.yaml) script been created that will spinup the application and its associated data stores and caching layers. With docker installed run the below script:

```sh
docker compose up
```

This server is listening on port `8080` by default. You can specify the required port you want by changing the `PORT` in the [`.env`](./.env) file to the required number and also changing the ports in the [`docker compose`](./docker-compose.yaml) script.

## Part 1

## i. Creating a random user

Once a running instance of the composed applications are running, in another shell instance you can send a simple `GET` request to the `/create/user` endpoint to get a randomly created user. You can test this by entering the following script into the second shell instance:

```sh
# be sure to change the port if you are using a custom port
curl localhost:8080/user/create
```

By using a command line JSON processing tool like [jq](https://stedolan.github.io/jq/) you can "pretty print" the output on your terminal as follows:

```sh
# be sure to change the port if you are using a custom port
curl localhost:8080/user/create | jq .
```

## ii. Logging in

Once a user has been created you can login by sending a `POST`request to `/login` with the following payload:

```
# be sure to change the email and password with the obtain from creating users above
curl -X POST \
-H "Content-Type: application/json" \
-d '{"email": "<string-changeme>", "password": "<string-changeme>"}' \
localhost:8080/login
```

## iii. Discovery

Once you have logged in you can find potential matches for the user by sending a `GET` request to `/discover` with the received token and the user Id in the header:

```
# be sure to change the Authorization with the token obtained from logging in above
curl -X GET \
-H "Content-Type: application/json" \
-H "Authorization: <string-changeme>" \
-H "Id: <integer-changeme>" \
localhost:8080/discover
```

## iv. Swiping

Users can swipe on each other to signify preference; this can be done by sending a `POST` request to `/swipe` with the received token in header:

```
# be sure to change the email and password with the obtain from creating users above
curl -X POST \
-H "Content-Type: application/json" \
-H "Authorization: <integer-changeme>" \
-d '{"swiperId": <integer-changeme>, "swipeeId": <integer-changeme>, "decision": <boolean-changeme>}' \
localhost:8080/swipe
```

## Testing

There has been a series of test cases that have been produced. This can be found in the [`service_test`](./service_test.go) file. 
These tests spin up a short-lived docker container and connects to it, runs manual migrations and tests the service through a 'real' and not mocked http request.
You can run the tests with the below command:

```sh
# -cover includes code coverage to the result
go test ./... -cover
```

Kind regards,

Ebuka Agbanyim.
ebuka7@outlook.com
