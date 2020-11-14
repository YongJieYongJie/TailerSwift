# TailerSwift
Command-line tool for filtering streaming JSON logs

## Installation
Run the following command in your favorite terminal to fetch the source from
this repositiory and compile the source:
```bash
> go get -u -ldflags="-s -w" github.com/yongjieyongjie/tailerswift
```
_Note: For information on the `-ldflags="-s -w"`, refer to this
[link](https://golang.org/cmd/link/#hdr-Command_Line)_.

## Usage

_Note: The examples below uses the provided `sample.log` as substitute for a
stream of logs on STDIN_.

- _Filter the JSON logs based a particular key_.   
  
  In the example below, only logs with `log_level` matching the regex `error`
  will be printed to STDOUT:
    ```bash
    > tail sample.log | tailerswift -filters='{\"log_level\": \"error\"}'
    
    # Output:
    {"log_level":"error","message":"Empty Space","request_id":"19d8be5a-c0bd-4a9b-b76b-fc44b0ddac59","timestamp":"2020-06-04T12:34:56.793+08:00"}
    {"log_level":"error","message":"You Belong With Her","request_id":"4b6a3a39-2761-4a50-9d4f-5a2d9231b9dc","timestamp":"2020-06-04T12:34:56.790+08:00"}
    ```

- _Print to STDOUT in CSV format_.  
  
  While JSON format is useful when moving data across systems, it is often
  more convenient to use CSV when commencing exploratory data analysis on the
  logs.

  The example below outputs the logs in CSV format (sorted by key,
  alphabetically).
    ```bash
    > tail sample.log | tailerswift -filters='{\"message\": \".*You.*\"}' \
    -outputFormat=csv

    # Output:
    info,Look What You Made Me Do,5a344536-fcdc-4752-811c-f8560fabb2e9,2020-06-04T12:34:56.791+08:00
    error,You Belong With Her,4b6a3a39-2761-4a50-9d4f-5a2d9231b9dc,2020-06-04T12:34:56.790+08:00
    ```

- _Projecting only certain fields in the output (similar to SQL's `SELECT`)_.  
  
  Good log messages should contain all the information necessary for
  debugging, but this also mean that they may be unnecessarily long for data
  analysis purposes.  
    
  Using the option `-project=<comma-separated-list-of-keys>`, we can output
  only the relevant fields. It is perfectly alright that the keys used
  for filtering are not projected in the output, the example below filters
  all log mesasges with `log_level` of `info`, but does not print the
  `log_level` to STDOUT.
  
    ```bash
    > tail sample.log | tailerswift -filters='{\"log_level\": \"info\"}' \
    -project="message"
    
    # Output:
    {"message":"Look What You Made Me Do"}
    {"message":"Love Story"}
    {"message":"ME!"}
    {"message":"Teardrops On My Guitar"}
    {"message":"Shake it Off"}
    {"message":"Style"}
    {"message":"22"}
    ```

- The number of goroutines to used is also configurable using
`-numGoroutines=<nunmber-of-goroutines-to-use>`:
    ```bash
    > tail sample.log | tailerswift -filters='{\"log_level\": \"info\"}' \
    -project="request_id,message" -outputFormat=csv -numGoroutines=4 
    ```

- As with other Go command-line tools using the built-in `flags` library, you
can get basic help using `<executable-binary> -help`:
    ```bash
    > tailerswift -help

    # Output:
    Usage of .../tailerswift:
      -filters string
            A JSON string containg key-regex pairs (default "{}")
      -numGoroutines int
            Number of concurrent goroutines to use for parsing the stream (default 4)
      -outputFormat string
            csv or json (default "json")
      -project string
            A comma-separated list of string representing keys to be printed out
    ```