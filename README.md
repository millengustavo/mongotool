# Go tool that fetch from API and insert on database for posterior Data Analysis

Get Transactions from API to a private MongoDB. Written in Go, the only external library necessary is mongo-go-driver, the official mongoDB driver for Go.

## Requirements

- Go
- Config variables (env)

## Config

The config parameters must be passed by environment variables:

```
export MongoURI="<URI>"

export AuthenticationAPI="<user>"

export AuthenticationKey="<key>"

export ReferenceID="?referenceId=<id>"

export LinkAPI="<linkapi>"

export LinkIntention="<linkintention>"

export DatabaseName="<dbname>"

export CollectionName="<collectionname>"
```
