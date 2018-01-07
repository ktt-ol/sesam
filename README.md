# Sesam

A member can login with his wiki credentials and activate our door buzzer. The activation is done by sending on the 'doorBuzzerTopic' topic.   


# Dependencies

Install all dependencies with
```
dep ensure
```


# Build

```
go build cmd/sesam.go
# or use the script 
./do.sh build-linux
```


# Config

Copy `config.example.toml` to `config.toml` and change as you like. 


# Run

```
./sesam
# or for production mode
GIN_MODE=release ./sesam 
```

You can also use the systemd service file `extras/sesam.service`


# TODO

* block a client after too many failed passwords attempts 