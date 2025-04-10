# Testing norm

## Basic Testing

```bash
make test
```

## Benchmark Testing Guide

QuerySet 包含了多种基准测试，用于测试不同查询构建操作的性能。以下是运行这些基准测试的方法及说明。

### Running All Benchmarks

To run all benchmark tests, use the following command:

```bash
make benchmark
```

or

```bash
go test -bench=. -benchmem
```

This will run all functions starting with `Benchmark` and display memory allocation statistics for each benchmark.

### Running Specific Benchmarks

If you only want to run specific benchmark tests, you can use regular expressions to match test names:

```bash
# Only run simple filter benchmarks
go test -bench=BenchmarkQuerySet_SimpleFilter -benchmem

# Run all benchmarks that include "Filter"
go test -bench=Filter -benchmem

# Run all benchmarks that include "Complete" or "Build"
go test -bench="(Complete|Build)" -benchmem
```

### Adjusting Benchmark Runtime

By default, Go automatically decides how many times to run each benchmark. You can adjust this using the `-benchtime` parameter:

```bash
# Run each benchmark for 3 seconds
go test -bench=. -benchtime=3s -benchmem

# Run each benchmark a specific number of times
go test -bench=. -benchtime=1000x -benchmem
```

### Comparing Performance Between Versions

You can use the `benchstat` tool to compare results from two benchmark runs:

```bash
# Run benchmarks with current version and save results
go test -bench=. -benchmem > old.txt

# Modify code, run benchmarks again, and compare
go test -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

### CPU and Memory Profiling

For deeper performance analysis, you can generate CPU and memory profile files:

```bash
# CPU profiling
go test -bench=BenchmarkQuerySet_ComplexFilter -cpuprofile=cpu.prof

# Memory profiling
go test -bench=BenchmarkQuerySet_ComplexFilter -memprofile=mem.prof

# Use pprof tool to view analysis results
go tool pprof cpu.prof
go tool pprof mem.prof
```

In the pprof interactive command line, you can use commands like `top`, `list`, and `web` to analyze performance bottlenecks.

### Benchmark Descriptions

- `BenchmarkQuerySet_SimpleFilter`: Tests the performance of simple filter conditions
- `BenchmarkQuerySet_ComplexFilter`: Tests the performance of complex filter conditions
- `BenchmarkQuerySet_MultipleFilters`: Tests the performance of multiple filter groups
- `BenchmarkQuerySet_Where`: Tests the performance of direct WHERE conditions
- `BenchmarkQuerySet_CompleteQuery`: Tests the performance of complete query building
- `BenchmarkQuerySet_BuildLargeQuery`: Tests the performance of large complex queries
- `BenchmarkQuerySet_FilterExclude`: Tests the performance of combined filter and exclude conditions
