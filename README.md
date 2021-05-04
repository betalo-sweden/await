[![Build Status](https://travis-ci.org/betalo-sweden/await.svg?branch=master)](https://travis-ci.org/betalo-sweden/await)

# This repository is no longer maintained.

Issue reports and pull requests will not be attended.
Please fork this repository if you wish to make changes.

# await

Await availability of resources.

This can be useful in the context of
[Docker Compose](https://docs.docker.com/compose/) where a service needs to wait
for other dependent services.

Optionally a timeout can be provided to specify how long to wait for all
dependent resources to become available. On success the command returns code
`0`, on failure it returns code `1`.

Additionally a command can be specified which gets executed after all dependent
resources became available.


## Installation

    go get -u github.com/betalo-sweden/await

or

    curl -s -f -L -o await https://github.com/betalo-sweden/await/releases/download/v0.4.0/await-linux-amd64
    chmod +x await


## Usage

    $ await -h
    Usage: await [options...] <res>... [ -- <cmd>]
    Await availability of resources.

      -V	Show version
      -f	Force running the command even after giving up
      -i string
        	Read resources from file, '-' to read from stdin
      -q	Set quiet mode
      -t duration
        	Set timeout duration before giving up (default 1m0s)
      -v	Set verbose output mode
      -vv
        	Set more verbose output mode


## Resources

All dependent resources must be specified as URLs or escaped command.

Some resources provided additional functionally encoded as fragment
(`#<fragment>`). The syntax follows the URL query syntax:
`k1|k1=|k1=v1,v2,v3...[&k2=v1&...]`.
E.g.: http://example.com/#ssl&foo=bar,baz&i=j

Valid resources are: HTTP, Websocket, TCP, File, PostgreSQL, MySQL, Command.


### HTTP Resource

**Availability**: Available when a connection to a given server is established
and an empty request returns the response status code is 2xx. Unavailable
otherwise.

**URL syntax**: `http[s]://[<user>[:<pass>]@]<host>[:<port>][<path>][?<query>][#<fragment>]`

**Fragment**:

- `tls=skip-verify`: When used, it skips TLS check for `https` resources.

### Websocket Resource

**Availability**: Available when a connection to a given server is established.
Unavailable otherwise.

**URL syntax**: `ws[s]://[<user>[:<pass>]@]<host>[:<port>][<path>][?<query>]`


### TCP Resource

**Availability**: Available when a connection to a given server is established.
Unavailable otherwise.

**URL syntax**: `tcp[4|6]://<host>[:<port>]`


### File Resource

**Availability**: Available when given path is a file and the file exists or, if
declared absent, the file does not exist. Unavailable otherwise.

**URL syntax**: `file://<path>[#<fragment>]`

- `absent` key: If present, the resource is defined as available, when the
  specific file is absent, rather than existing.


### PostgreSQL Resource

**Availability**: Available when a connection to a given PostgreSQL database
server is established and optional a given database could be found and optional
any tables or a set of tables could be found. Unavailable otherwise.

**URL syntax**: `postgres://[<user>[:<pass>]@]<host>[:<port>][/<dbname>][?<dbparams>][#<fragment>]`

The URL defines a [DSN](https://en.wikipedia.org/wiki/Data_source_name).

The database name `<dbname>` is optional. If provided, the resource is
classified as available as soon as the database was found.

**DB Parameters**:

- `sslmode=[verify-ca|require]`: `sslmode=verify-ca` enables TLS/SSL encrypted
  connection to the server. Use `sslmode=require` if you want to use a
  self-signed or invalid certificate (server side). See
  [lib/pq](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters)
  for more details.

**Fragment**:

- `tables[=t1,t2,...]` key-value: If key present and value absent, the
  resource's database scheme must at least contain one table. If key present and
  value present, the resource's database scheme must at least contain the
  specified tables. Using this key requires to provide a database name.


### MySQL Resource

**Availability**: Available when a connection to a given MySQL database server
is established and optional a given database could be found and optional any
tables or a set of tables could be found. Unavailable otherwise.

**URL syntax**: `mysql://[<user>[:<pass>]@]<host>[:<port>][/<dbname>][?<dbparams>][#<fragment>]`

The URL defines a [DSN](https://en.wikipedia.org/wiki/Data_source_name).

The database name `<dbname>` is optional. If provided, the resource is
classified as available as soon as the database was found.

**DB Parameters**:

- `tls=[true|skip-verify]`: `tls=true` enables TLS/SSL encrypted connection to
  the server. Use `tls=skip-verify` if you want to use a self-signed or invalid
  certificate (server side). See
  [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql#tls) for more
  details.

**Fragment**:

- `tables[=t1,t2,...]` key-value: If key present and value absent, the
  resource's database scheme must at least contain one table. If key present and
  value present, the resource's database scheme must at least contain the
  specified tables. Using this key requires to provide a database name.

### AMQP resource

**Availability**: Available when a connection to the given AMQP server is established. Unavailable otherwise.

**URL syntax**: `amqp[s]://[<user>[:<pass>]@]<host>[:<port>]/[<vhost>]`

### Command Resource

Does not follow the URL syntax and is used as generic fallback for invalid URLs
(i.e. schema is absent).

**Availability**: Available when command return status code `0`. Unavailable
otherwise.

**URL syntax**: `<path> [<arg>...]`


## Alternatives

Once Docker-Compose support the Docker `HEALTHCHECK` directive for awaiting
services being up, this tool could become deprecated.

Many alternative solutions exist, most of them avoiding to interpret the
resource type, leaving only options like tcp and http, or having a specific
focus on Docker Compose. A few of them listed below:

- [ContainerPilot](https://github.com/joyent/containerpilot)
- [controlled-compose](https://github.com/dansteen/controlled-compose)
- [crane](https://github.com/michaelsauter/crane)
- [docker-wait](https://github.com/aanand/docker-wait)
- [dockerize](https://github.com/jwilder/dockerize)
- [wait-for-it](https://github.com/vishnubob/wait-for-it)
- [wait_for_db](https://gitlab.com/thelabnyc/ci/blob/09504268779acf53d65383b98b76e44ff763ef4d/examples/docker-compose-links/entrypoint.sh)
- [wait_to_start.sh](https://gist.github.com/rochacbruno/bdcad83367593fd52005#file-wait_to_start-sh)
- [waitforservices](https://github.com/Barzahlen/waitforservices)


## License

MIT License. See [LICENSE](LICENSE).
