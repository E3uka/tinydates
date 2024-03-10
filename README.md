# Tinydates
Muzz technical assessment by Ebuka Agbanyim

## How to run
This application is built and deployed user docker,
[`a docker compose script`](./docker-compose.yaml) been created that will spinup the application and its associated data stores and caching layers. With docker installed run the below script:

```sh
docker compose up
```

This server is listening on port `8080` by default. You can specify the required port you want by changing the `PORT` in the [`.env`](./.env) file to the required number and also changing the ports in the [`a docker compose script`](./docker-compose.yaml).

## Part 1

## i. Creating a random user

Once a running instance of the composed applications are running, in another shell instance you can send a simple `GET` request to the `/create/user` endpoint to get a randomly created user. You can test this by entering the following script into the second shell instance:

```sh
# be sure to change the port if you are using a custom port
curl localhost:8080/user/create
```

By using a command line JSON processing tool like [jq](https://stedolan.github.io/jq/) you can
"pretty print" the output on your terminal as follows:

```sh
# be sure to change the port if you are using a custom port
curl localhost:8080/user/create | jq .
```

## ii. Logging in

Once a user has been crated you can login by sending a `POST`request to `/login` with the following payload:

```
{
  "email": "[user email here]",
  "password": "[user password here]",
}
```

you can run the following to make the login request test:

```
# be sure to change the email and password with the obtain from creating users above
curl -X POST -H "Content-Type: application/json" -d '{"email": "changeme", "password": "changeme"}' localhost:8080/login
```

## Testing

There has been a series of test cases that have been produced. This can be found in the
[`service_test`](./service_test.go) file. You can run the tests with the below command:

```sh
# -cover includes code coverage to the result
go test ./... -cover
```

Kind regards,

Ebuka Agbanyim.
ebuka7@outlook.com
