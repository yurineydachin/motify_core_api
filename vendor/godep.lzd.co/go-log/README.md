# go-log
Official logger library for Go services

- [Usage](#Usage)
- [Installation](#Installation)

## Usage

```go
package main

import (
    "godep.lzd.co/go-log"
    "godep.lzd.co/go-log/format"
)
    
// create Logger
l = log.NewLogger("service_name", "unixgram", "/dev/log", log.DEBUG)

// write in Standart log format
l.Write(log.DEBUG, &format.Std{
    TraceId: "6yaoivssj1rt"
    SpanId: "6yaoivssj1rt",
    ParentSpanId: "3ymrswshj4sg",
    Message: "test message",
    Data: map[string]interface{}{"attribute":"value"}
})

// write in CEE format
l.Write(log.DEBUG, &format.CEE{
    Data: "any data",
})

// change log level
l.SetLevel(log.ERROR)
```

## Formats

- Standart log format (`format.Std`)
```
<Priority> {YYYY-mm-ddTHH:mm:ss.microseconds±hh:mm} | {TraceId} | {ParentSpanId} | {SpanId} | {rollout_type} | {service_name} | {level} | {component_name} | {file_name_and_line} | {message} | {additional_data} | {is_truncated}
```
- JSON Messages over Syslog (`format.CEE`)
```
<Priority> @cee: {JSON}
```


## Installation

    go get godep.lzd.co/go-log

## Benchmark

    BenchmarkLogWithoutNet-4                         5000000              3312 ns/op            1091 B/op         10 allocs/op
    BenchmarkSyncPoolLogWithoutNet-4                 5000000              2714 ns/op             547 B/op          7 allocs/op
    BenchmarkLeakyBufferLogWithoutNet-4              5000000              2760 ns/op             547 B/op          7 allocs/op
    BenchmarkLogWithNet-4                            1000000             10520 ns/op            1091 B/op         10 allocs/op
    BenchmarkSyncPoolLogWithNet-4                    1000000             10062 ns/op             547 B/op          7 allocs/op
    BenchmarkLeakyBufferLogWithNet-4                 1000000             10479 ns/op             547 B/op          7 allocs/op
    BenchmarkLogWithoutNetParallel-4                10000000              1697 ns/op            1091 B/op         10 allocs/op
    BenchmarkSyncPoolLogWithoutNetParallel-4        10000000              1469 ns/op             547 B/op          7 allocs/op
    BenchmarkLeakyBufferLogWithoutNetParallel-4     10000000              1510 ns/op             547 B/op          7 allocs/op
    BenchmarkLogWithNetParallel-4                    1000000             19864 ns/op            1092 B/op         10 allocs/op
    BenchmarkSyncPoolLogWithNetParallel-4            1000000             18740 ns/op             548 B/op          7 allocs/op
    BenchmarkLeakyBufferLogWithNetParallel-4         1000000             18976 ns/op             548 B/op          7 allocs/op