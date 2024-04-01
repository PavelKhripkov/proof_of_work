# PROOF OF WORK

### About the project
Current project is a Go client-server application that is using one of 
Proof Of Work algorithms designed to protect server from DoS-like attacks
and that requires from the client to do some work before they can access
the server's resources.
Implementation meets the requirements of the following test task.

### Original task text:
Design and implement “Word of Wisdom” tcp server.
* TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
* The choice of the POW algorithm should be explained.
* After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
* Docker file should be provided both for the server and for the client that solves the POW challenge.

### Proof of work algorithm
The algorithm used in the current implementation is a modification of Hashcash.
It was chosen among others because this algorithm is most-referenced when we're googling for PoW.
This algorithm is used in Bitcoin. And this is the first implementation of such kind of algorithms by the author.

### App architecture
Application has classic architecture for Go projects. There are server service and client service.
Redis is used as a storage because of it's TTL feature. Hashcash implemented in a separate package,
so it can be imported into other projects and reused.

There are key features of Hashcash implementation:
* Concurrent searching for nonce. User of package can choose concurrency factor.
* In-place nonce changing while searching for correct one. Significantly decreases consumed resources and increases searching speed.
* User can choose the hash function to be used. But important that client uses the same one as the server.

### Requirements
To run in container:
* Docker v.24.0.7 (or other supporting multi-stage build syntax)
* Go v1.22

For local run additionally need:
* Redis v7.2.4 (should work from v6.2.0)

### Run server
There is a docker compose file that can be used to run server. 
But the easiest way is to type:
```sh
make server-up
```
To stop the server and remove containers type:
```sh
make server-down
```

### Run client
After we have the server running, we can run the client. Type:
```sh
make client-up
```
The client makes requests to the server with different settings.
First client properly finishes the required job and sends results to the server,
getting some wise quotes as a reward. Then settings are changed, and we get
error responses with different codes from the server. And finally we won't give
the client enough time to finish the job, and the timeout error will be the only
thing we can get in that case.

Type the following to remove client container from the system:
```sh
make client-down
```

### Run tests
There are tests and linters runner in the [Makefile](Makefile). 
First prepare the tools and then run both of them with a single command:
```sh
make init
```
```sh
make test
```
We encounter the linter's complaints, because there are no linters disabled, not even annoying ones. 
But it's possible to disable some of the in the [.golangci.yml](.golangci.yml) file