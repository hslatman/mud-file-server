# MUD File Server

A MUD File Server using [Caddy](https://caddyserver.com/) as the web server.

## Description

[Manufacturer Usage Descriptions](https://tools.ietf.org/html/rfc8520) (MUDs) allow manufacturers of IoT equipment to specify the intended network communication patterns of the devices they manufacture. 
The access control policies described in a MUD File allow network controllers to automatically enforce rules on the device, resulting in devices only being allowed to communicate within the boundaries of the access control policies.
MUD Files typically don't reside on the (local) network itself, which is why MUD Controllers need a way to retrieve MUD Files as soon as they find out that a MUD File exists for a device.
MUD File Servers are responsible for serving MUD Files and their signatures, which can be retrieved by MUD Controllers when an IoT device emits a MUD URL.

MUR URLs have the following basic properties:

* They always use the "https" scheme
* Any "https://" URL can be a MUD URL

This repository contains an implementation of a MUD File Server based on the [Caddy](https://caddyserver.com/) web server.
It is implemented as a Caddy module and can thus be embedded in a Caddy deployment like any other module.
The module is available to be imported as follows:

```go
import (
    _ "github.com/hslatman/mud-file-server/pkg/mud".
)
```

The repository also contains an example command for running a Caddy server with the MUD File Server enabled.
This can be found in the `cmd` directory.

## Build

The MUD File Server binary can be built as shown below:

```bash
# build the server 
$ go build cmd/main.go -o muds
```

## Usage

```bash
# run the server directly from Go code, using the provided config.json 
$ go run cmd/main.go run --config config.json
# run the server from compiled binary, using the provided config.json 
$ ./muds run --config config.json
```

Now the MUD File Server can be reached at https://localhost:9443/.
Assuming the examples directory is available and the repository directory set as the root to serve files from, the example MUD file for `The BMS Example Light Bulb` can now be retrieved from:

https://localhost:9443/examples/lightbulb2000.json

And its can be retrieved from signature from:

https://localhost:9443/examples/lightbulb2000.json.p7s

Files that are invalid MUD Files or not parseable as CMS objects are not served by default.

### Configuration

The MUD File Server module can be configured like any other Caddy module.
We've provided a sample config.json file, which sets the root of the MUD File Server to be the current directory and disables request header validation (for demo purposes only).
The following options are available:

* `root`: string that indicates the root directory to serve MUD Files and signatures from. Defaults to the Caddy `{http.vars.root}` parameter if set, but current working directory otherwise.
* `validate_headers`: boolean that indicates requests headers should be validated or not. Disabling this is easier for demos in a web browser, but should not be done for an actual server. Defaults to true.
* `validate_mud`: boolean that indicates the contents of a MUD file should be validated or not. Validation is performed using https://github.com/hslatman/mud.yang.go/. The signature is NOT validated. Defaults to true.
* `set_etag`: boolean that indicates whether or not to set the ETag header in responses. Defaults to true.

## Signing & Verifying MUD Files

According to RFC 8520, MUD files MUST be signed using Cryptographic Message Contents (CMS).
Within the MUD file itself, the `mud-signature` property points to the location where the (detached) signature can be found.
The `mud-signature` property can be used by a MUD Manager to retrieve the signature file.
By default, the assumption is that the location of the signature file is right next to the MUD file itself, but it can be somewhere different.
A small caveat is that the location of the signature file should be set before signing the MUD file, because it's in the contents of the MUD file to be signed.

Signing a MUD file is described in the [RFC](https://tools.ietf.org/html/rfc8520#section-13).
An example command invocation looks like this:

```bash
# within the mud-file-server repository, assuming example certificates and keys are available
$ openssl cms -sign -signer cert/server.crt -inkey cert/server.key -in examples/lightbulb2000.json -binary -outform DER -binary -certfile cert/intermediate.crt -out examples/lightbulb2000.json.p7s
```

Signatures can be checked as follows:

```bash
# within the mud-file-server repository, assuming the certificate that was used for signing the file is trusted
$ openssl cms -verify -in examples/lightbulb2000.json.p7s -inform DER -content examples/lightbulb2000.json
```

### Example Certificates

A small utility script for generating the keys and certificates for signing a MUD File has been included in this repository.
It serves as an example; it probably shouldn't be used as is for production deployments.
It can be used as follows:

```bash
# within the cert directory:
$ ./new.sh
```

The certificate and key generated are directly under the newly generated CA, so no intermediates are included.
The command for signing a MUD File using a key and certificate generated with the utility script should thus be changed to not include the intermediate certificate like below:

```bash
# within the mud-file-server repository, assuming example certificates and keys are available
$ openssl cms -sign -signer cert/server.crt -inkey cert/server.key -in examples/lightbulb2000.json -binary -outform DER -binary -out examples/lightbulb2000.json.p7s
```

## Goal

The main goal of this repository is to provide a reference MUD File Server implementation that is compliant with RFC 8520.
By choosing Caddy as the server, which has been steadily growing in popularity because of its ease of deployment and automatic TLS configuration, it might become easier for companies to deploy a MUD File Server.

Caddy was on my list of things to learn about and work with, so this little project allowed me to do just that.

## TODO

* Add logging using Caddy provided logger?
* Do we need some kind of abstract file system handling?
* Implement a simple overview page of MUDs available?
* Implement a MUD viewer to visualize available MUDs?
* Implement basic statistics about files requested?
* More robust content type checking?
* Add commands for signing / verifying MUD signatures? Or should that be part of mud.yang.go?
* Add signature verification before serving the MUD file (if signature exists; on same server, or a different one)?
* Add non-Caddy implementation to be embedded in any Go project?