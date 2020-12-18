# HybridLog
HybridLog is an Golang implementation of an append-only file with concurrent reads/writes support and durability guarantee. 

HybridLog writes as fast as `os.File.Write` with `O_APPEND` flag, but performs read ~20x faster than `os.File.ReadAt` on 512b-blocks, ~10x faster on 4KB-blocks
and ~1.2x faster on 1M-blocks.

## Getting Started
### Installation
```
go get github.com/snowflk/hybridlog/...
```
### Create a hybrid log
```go
package main
import (
    "log"

    "github.com/snowflk/hybridlog"
)

func main()  {
    hlog, err := hybridlog.Open(hybridlog.Config{
        Path:"mydata.log",
    })
    if err != nil{
        log.Fatal(err)    
    }
    defer hlog.Close()
    // Do something
}
```

### Write
HybridLog implements the interface `io.Writer`, so you can use it as follows:
```go
if _, err := hlog.Write([]byte("My data")); err != nil {
    return err
}
```
### Read
HybridLog implements the interface `io.ReaderAt`, so you can use it as follows:
```go
// Read 128 bytes
data := make([]byte, 128)
if _, err := hlog.ReadAt(data, 0); err != nil {
    return err
}
```
## Advanced configurations
### Auto Compaction
Each write operation creates a fragment. Too fragmented data will result to negative impact on performance. AutoCompaction takes care
of the defragmentation process.

Open HybridLog in AutoCompaction mode to enable this feature. There are 2 modes: `TimeBased` and `FragmentationBased`.

`TimeBased` mode performs compaction in a fixed interval. 
`FragmentationBased` mode performs compaction when the number of fragments exceeds a defined threshold.
```go
hlog, err := hybridlog.Open(hybridlog.Config{
    Path:"mydata.log",
    AutoCompaction: true,
    CompactionMode: hybridlog.TimeBased,
    CompactAfter: 15 * time.Minute // Perform compaction every 15 minutes
    // ...
})
```
### Sync Policy
In order to guarantee durability, fsync must be called. However, fsync has negative impact on write performance.
Therefore, you can choose a sync policy to configure when to perform fsync, depending on your requirements.

There are 3 policies: 
- `NoSync` let the system determine when to sync. This is the default policy.
- `AlwaysSync` for strong durability, this policy uses the flag O_SYNC for opening the file.
- `SyncEverySec` performs sync every second. 
```go
hlog, err := hybridlog.Open(hybridlog.Config{
    Path:"mydata.log",
    SyncPolicy: hybridlog.AlwaysSync,
    // ...
})
```

## Benchmark
In this benchmark, HybridLog will be compared to the built-in `os.File`.


#### Technical specification
Processor: 2,8 GHz Quad-Core Intel Core i7  
RAM: 16 GB 2133 MHz LPDDR3  
Disk: Apple built-in SSD  

#### HybridLog's Result
```

BenchmarkHybridLog_Write_512b
BenchmarkHybridLog_Write_512b-8    	  116798	      9645 ns/op
BenchmarkHybridLog_Write_1KB
BenchmarkHybridLog_Write_1KB-8     	  107182	     11045 ns/op
BenchmarkHybridLog_Write_4KB
BenchmarkHybridLog_Write_4KB-8     	   60058	     19029 ns/op
BenchmarkHybridLog_Write_128KB
BenchmarkHybridLog_Write_128KB-8   	   10000	    131224 ns/op
BenchmarkHybridLog_Write_1MB
BenchmarkHybridLog_Write_1MB-8     	    1600	    976735 ns/op
BenchmarkHybridLog_Read_512b
BenchmarkHybridLog_Read_512b-8     	21326368	        54.8 ns/op
BenchmarkHybridLog_Read_1KB
BenchmarkHybridLog_Read_1KB-8      	19704802	        61.0 ns/op
BenchmarkHybridLog_Read_4KB
BenchmarkHybridLog_Read_4KB-8      	13594483	        89.3 ns/op
BenchmarkHybridLog_Read_128KB
BenchmarkHybridLog_Read_128KB-8    	  292083	      3861 ns/op
BenchmarkHybridLog_Read_1MB
BenchmarkHybridLog_Read_1MB-8      	   24596	     49447 ns/op
```
#### os.File's Result
```
BenchmarkFile_Write_512b
BenchmarkFile_Write_512b-8         	  126115	      9682 ns/op
BenchmarkFile_Write_1KB
BenchmarkFile_Write_1KB-8          	  101589	     11216 ns/op
BenchmarkFile_Write_4KB
BenchmarkFile_Write_4KB-8          	   99919	     11087 ns/op
BenchmarkFile_Write_128KB
BenchmarkFile_Write_128KB-8        	   12734	     97205 ns/op
BenchmarkFile_Write_1MB
BenchmarkFile_Write_1MB-8          	    1420	    789027 ns/op
BenchmarkFile_Read_512b
BenchmarkFile_Read_512b-8          	 1325258	       916 ns/op
BenchmarkFile_Read_1KB
BenchmarkFile_Read_1KB-8           	 1333854	       922 ns/op
BenchmarkFile_Read_4KB
BenchmarkFile_Read_4KB-8           	 1205302	       981 ns/op
BenchmarkFile_Read_128KB
BenchmarkFile_Read_128KB-8         	  182815	      6725 ns/op
BenchmarkFile_Read_1MB
BenchmarkFile_Read_1MB-8           	   18992	     60004 ns/op
```