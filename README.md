# Bank Integration Library

This library provides tools to integrate with various banking services in Indonesia.
For now, this library only supports BCA. Other bank will be added in the future.

## Installation

To install the library, use `go get`:

```sh
go get github.com/voxtmault/bank-integration
```

## Importing

To import the library in your Go project, add the following import statement:

```go
import "github.com/voxtmault/bank-integration"
```

## Usage

Here is a basic example of how to use the library:

```go
package main

import (
    "fmt"
    bi "github.com/voxtmault/bank-integration"
)

func main() {
    client := bi.InitBankAPI("banking.env", <TimeZone>)
    bcaService, err := bankService.InitBCAService()
    if err != nil {
       return eris.Wrap(err, "failed to initialize BCA service")
    }
    fmt.Println("Account:", account)
}
```

## Requirement

This library requires a database account that has sufficient permission to Create, Read, and Update data into multiple tables. Optionally, you can add permission to create new tables that is going to be used to log http request coming from and going to external bank services.

## Database Migration

All the required sql scripts are provided in the [Migration](db/changelog/) folder. You can run some kind of database migration tools (i use liquibase) to apply the initial db structure or you can run them manually to add the tables to your existing database. I'd reccomend to separate your main program database and payment database.
