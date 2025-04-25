# HAFH-Server

Your "Home Away From Home" server for an at-home view of your network-enabled, MQTT-driven devices.

## Synopsis

In the simplest terms, this project aims to provide an easy-to-use interface for devices on your home network to report their data (readings) to a central location that can be queried by other devices or applications.

This project provides a simple executable that contains both an HTTP server and an MQTT broker. The HTTP server is used to retrieve information about previously-connected peripherals, while the MQTT broker is used to receive messages from the peripherals. The application is designed to be run on a Raspberry Pi 02W or similar device, but it can also be run on any other device that supports Go.

## Features

- MQTT Broker (TLS)
- HTTP Server (HTTPS via `ngrok`)
- SQLite Database for secure and persistent storage
- Configurable via YAML

## Getting Started

### Dependencies

This project is built with [Go](https://go.dev/doc/install) and has one outside Go dependency for linting purposes:

```sh
go install honnef.co/go/tools/cmd/staticcheck@latest
```

The only other dependency is `make`, which is typically pre-installed on most Unix-like systems. If you don't have `make`, you can install it using your system's package manager.

### Configuration

The configuration for the application is stored in a YAML file that is parsed at runtime. For examples of the configuration file, see the [`configs/` directory](./configs/). **The configuration file is required to run the application and must be passed to the resulting executable as a file path.**

### `make`

This project uses `make` to manage the build, formatting, linting, and other tasks. See the [`Makefile`](./Makefile) for the available targets. The most common targets are:

- `make certs`: Generate self-signed certificates for the MQTT broker.
- `make build`: Build the application.
- `make dev`: Build and run the development application.
- `make lint`: Run the linter on the codebase.
- `make format`: Format the codebase.

## HTTP

The HTTP server is a simple REST API that allows callers to retrieve information about reporting peripherals.

### Endpoints

The HTTP server exposes the following endpoints:

- `GET /api/v1/version`: Returns the API version.
- `GET /api/v1/peripherals`: Returns a list of all the previously-connected peripherals.
- `POST /api/v1/peripherals`: Sets the name and type of a peripheral. The body of the request should be a JSON object with the following fields:
  - `serial_number`: The serial number of the previously-connected peripheral.
  - `name`: The new name of the peripheral.
  - `type`: The integer representing the type (see `PeripheralType` in [database.go](./internal/database/database.go) for the list of types).
- `POST /api/v1/readings`: Gets up-to the specified number of readings of the requested peripheral. The body of the request should be a JSON object with the following fields:
  - `serial_number`: The serial number of the peripheral.
  - `num_readings`: The maximum number of readings to return.

### HTTP Authentication

Authentication for the HTTP server is done using a simple API key. The API key is passed in the `X-API-Key` header of the request. The key is stored in the configuration file and is required to access any of the endpoints.

Example:

```sh
curl -X GET \
  http://localhost:8080/api/v1/version \
  -H 'X-API-Key: <your-api-key>'
```

### HTTPS with `ngrok`

As mentioned above, this application exposes a standard HTTP server. _However_, for the HTTP server to have any use, we need to expose it to the outside world. This can be done using `ngrok`, which creates a secure (HTTPS) tunnel to your localhost. To do this, create an account (or sign in) and follow the [getting started guide](https://dashboard.ngrok.com/get-started) before following one of the two options below.

>**Note**: it is HIGHLY recommended to create a static domain for your `ngrok` tunnel. One static domain is free, but you can create more for a small fee. This will allow you to use the same domain every time you start the tunnel.

#### Internal / Embedded (Recommended)

The application can optionally be built with `ngrok` embedded. This is done by configuring the `ngrok` settings in the configuration file. The application will then automatically start an `ngrok` tunnel when it starts up. This is the recommended way to run the application, as it simplifies the setup process and ensures that the tunnel is always running.

>**Note**: This option requires a static domain be set in the configuration file.

#### External

Once you have `ngrok` installed, you can run the following command to expose your HTTP server:

```sh
ngrok http 8080
```

Or, to start the `ngrok` tunnel with your static domain, run:

```sh
ngrok http --url=<your-domain> http://localhost:8080
```

## MQTT

This application also includes an MQTT broker that is used to receive messages from the peripherals. The broker is configured to listen on port `8883` by default, but this can be changed in the configuration file. Although the broker is intended for peripheral reporting, it can also be used as a general-purpose MQTT broker between clients.

### MQTT Topics & Adding Peripherals

The MQTT broker will handle all messages as needed but specifically listens to `/peripherals/readings/#`. This topic is used to receive readings from peripherals, where `#` is a wildcard that matches any number of subtopics. The readings are expected to be in JSON format and should include the following fields:

- `serial_number`: The serial number of the peripheral.
- `data`: The JSON object containing the reading data, intended to be "dumb" (i.e., no processing is done on the data - it is the HTTP API caller's responsibility to interpret the data).
- (Optional) `timestamp`: The timestamp of the reading in ISO 8601 format.

Any readings received are stored in the SQLite database for later retrieval via the HTTP API. A few important notes:

- If the reading is not in the expected format, it will be ignored and logged as an error
- **If a reading comes in from a peripheral that is not registered, the peripheral will first be created in the database with the serial number and type set to `0` (which can be updated later via the HTTP API)**

### MQTT Authentication

The MQTT broker is configured to use mTLS for authentication. This means that both the client and server must present a valid certificate to establish a connection. The certificates are generated using the `make certs` command, which creates self-signed certificates for the server and client. The server certificate is used to authenticate the server, while the client certificate is used to authenticate the client.
