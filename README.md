# dora - bmc discovery asset database

A tool to build a dynamic database of you datacenter assets

## How to run

Small installations / Dev Setup:

### Docker

```console
git clone github.com/bmc-toolbox/dora
cd dora
# build docker image with application
docker-compose build dora
# start server in background, accessable by address http://localhost:8000
docker-compose up -d server
# run dora commands
docker-compose run dora scan 192.168.0.0/24
docker-compose run dora collect
```

### Outside of Docker

```console
git clone github.com/bmc-toolbox/dora
cd dora
go build -tags="gingonic" -ldflags="-s -w"
# start server, accessable by address http://localhost:8000
./dora --config dora-simple.yaml server
# run dora commands
./dora scan 192.168.0.0/24
./dora collect
```

Kea example configuration file to load subnets can be found by name
 [kea-simple.conf](kea-simple.conf).

## Requirements

Database - any compatible with [GORM](http://gorm.io/)

## Overview

Dora is a service which gather data about database assets from BMCs
 via HTTP\SSH and give ability to retrieve that data via REST API.

### Architecture

#### Server

Dora web server provides API for querying the data which Dora collected.

#### Worker

Worker consumes jobs issued by `publish` command, perform them and
 write results to database.

There are two type of jobs:

* `collect`: collects hosts found by the scanner or collect a given list of hosts
 (fast operation, except for Dell servers)
* `scan`: scan networks found in kea config or a list of given networks (slow operation)

In case you run these jobs as commands to dora, it works as a worker who received 
the command.

## Acknowledgment

dora was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.
