# Web API/UI for exploring the network

The block explorer is an HTTP daemon serving an API for exploring the CacheCash
network.

## Quickstart

Browse to [block explorer](https://explorer.testnet.cachecashcdn.net/).

## API

The explorer performs media detection; use postman or similar browser tools to
perform API queries directly.

Supported media types:

* application/json
* application/protobuf
* text/html

API:
(planned)

* /blocks
* /blocks/ids/`<id>`
* /caches
* /caches/ids/`<id>`
* /escrows - filter
    tickets redeemed by cache?
* /publishers
* /publishers/ids/`<id>`

API traversal patterns - start with blocks
blocks lead to escrows
escrows lead to caches and publishers
show selected metrics?
show tickets for escrows?
