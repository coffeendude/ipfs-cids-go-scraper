# My IPFS Scraper Go Application

This application is a Go-based service that interacts with a PostgreSQL database and the InterPlanetary File System (IPFS). It fetches and stores metadata for Content Identifiers (CIDs) from IPFS into a PostgreSQL database. The application also provides an HTTP server that allows users to fetch the stored metadata.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Running the Application](#running-the-application)
- [Benchmark Tests](#benchmark-tests)
- [Profiling](#profiling)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

- Go (version 1.21 or later)
- PostgreSQL (version 13 or later)

## Installation

```bash
git clone https://github.com/coffeendude/ipfs-cids-go-scraper.git
cd ipfs-cids-go-scraper
go get ./...
```

## Running the Application
To run the application, you need to have Go installed on your machine. You can then clone the repository and run the main file:

`go run main.go`

The application accepts several command-line flags to configure the database connection:

-host: The database host (default: "localhost")

-port: The database port (default: "5432")

-user: The database user (default: "postgres")

-password: The database password (default: "example")

-dbname: The database name (default: "postgres")


-sslmode: The SSL mode (default: "disable")

For example, to connect to a database on localhost:5432 as the user postgres with the password example, you would run:

`go run main.go -host=localhost -port=5432 -user=postgres -password=example -dbname=postgres -sslmode=disable`

**Note: For connecting to a remote AWS RDS instance, you need to set -sslmode='require'

## Benchmark Tests
The application includes two benchmark tests: BenchmarkFetchAndStoreMetadata and BenchmarkConnectToDB. These tests measure the performance of fetching and storing metadata for a list of CIDs in the database, and connecting to the database, respectively. These tests are important for understanding the performance characteristics of the application and identifying potential bottlenecks.

To run the benchmark tests, use the go test command with the -bench flag:

`go test -bench=.`

The output of the benchmark tests shows the number of iterations that were run and the average time per operation. For example:

```bash
BenchmarkFetchAndStoreMetadata-4   	    2000	    138344184 ns/op
BenchmarkConnectToDB-4             	    1000	    123456789 ns/op
```

In this example, the BenchmarkFetchAndStoreMetadata test ran 2000 iterations, and each operation took an average of 138344184 nanoseconds (or about 0.138 seconds). The BenchmarkConnectToDB test ran 1000 iterations, and each operation took an average of 123456789 nanoseconds (or about 0.123 seconds).

Remember, the actual numbers will vary depending on your hardware, the size of the ipfs_cids.csv file, and other factors. Always measure performance in a controlled environment to ensure accurate results.

Additionally, the application includes the use of the go pprof tool that enables the developer to collect profiling data like CPU and memory

To use the go tool pprof, append the desired profiling flag to the above test command"
Example:

`go test -bench=. -cpuprofile=cpu.out` 

or

`go test -bench=. -memprofile=mem.out`

And you can use the go tool pprof command to analyze the memory profiling data:

`go tool pprof mem.out`