# MUD File Server

A MUD File Server using [Caddy](https://caddyserver.com/) as the web server.

## Description

[Manufacturer Usage Descriptions](https://www.rfc-editor.org/rfc/rfc8520) (MUDs) allow manufacturers of IoT equipment to specify the intended network communication patterns of the devices they manufacture. 
The access control policies described in a MUD file allow network controllers to automatically enforce rules on the device, resulting in devices only being allowed to communicate within the boundaries of the access control policies.

This repository contains an implementation of a MUD File Server based on the Caddy web server.
MUD File Servers are responsible for serving MUD Files and their signatures, which can be retrieved by MUD Controllers when an IoT device emits a MUD URL.

MUR URLs have the following basic properties:

* They always use the "https" scheme
* Any "https://" URL can be a MUD URL

## Build

Because the repository is currently private, we need to instruct Go that the module is private when trying to build a version of running the application:

```bash
# build the server 
$ GOPRIVATE="github.com/hslatman/mud-file-server" go build cmd/main.go -o muds
```

## Usage

```bash
# run the server directly from Go code
$ GOPRIVATE="github.com/hslatman/mud-file-server" go run cmd/main.go
# run the server from compiled binary
$ ./muds
```

## TODO

* Add logging
* Change from Path Matcher to File Matcher?
* Test TLS (local CA?)
* Add example MUD including signature
* Do we need some kind of abstract file system handling?
* Do we need a simple database?
* Implement a simple overview page of MUDs available?
* Implement a MUD viewer to open the visualize the available MUDs?
* Implement basic statistics about files requested?
* More robust content type checking?
* Decrease the Caddy footprint by importing only what we use