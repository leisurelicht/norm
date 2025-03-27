# norm
# test command

```bash
make test
```

# 基准测试使用说明

QuerySet 包含了多种基准测试，用于测试不同查询构建操作的性能。以下是运行这些基准测试的方法及说明。

## 运行所有基准测试

要运行所有基准测试，可以使用以下命令：

```bash
make benchmark
```
or 

```bash
go test -bench=. -benchmem
```

这将运行所有以`Benchmark`开头的函数，并显示每个基准测试的内存分配统计信息。

## 运行特定的基准测试

如果只想运行特定的基准测试，可以使用正则表达式来匹配测试名称：

```bash
# 只运行简单过滤器的基准测试
go test -bench=BenchmarkQuerySet_SimpleFilter -benchmem

# 运行所有包含Filter的基准测试
go test -bench=Filter -benchmem

# 运行所有包含Complete或Build的基准测试
go test -bench="(Complete|Build)" -benchmem
```

## 调整基准测试运行时间

默认情况下，Go会自动决定运行每个基准测试的次数。你可以使用`-benchtime`参数来调整：

```bash
# 每个基准测试运行3秒
go test -bench=. -benchtime=3s -benchmem

# 每个基准测试运行特定次数
go test -bench=. -benchtime=1000x -benchmem
```

## 比较不同版本的性能

可以使用`benchstat`工具比较两次基准测试的结果：

```bash
# 运行当前版本的基准测试并保存结果
go test -bench=. -benchmem > old.txt

# 修改代码后运行基准测试并比较
go test -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

## CPU和内存分析

要进行更深入的性能分析，可以生成CPU和内存分析文件：

```bash
# CPU分析
go test -bench=BenchmarkQuerySet_ComplexFilter -cpuprofile=cpu.prof

# 内存分析
go test -bench=BenchmarkQuerySet_ComplexFilter -memprofile=mem.prof

# 使用pprof工具查看分析结果
go tool pprof cpu.prof
go tool pprof mem.prof
```

在pprof交互式命令行中，可以使用`top`、`list`、`web`等命令来分析性能瓶颈。

## 基准测试说明

- `BenchmarkQuerySet_SimpleFilter`: 测试简单过滤条件的性能
- `BenchmarkQuerySet_ComplexFilter`: 测试复杂过滤条件的性能
- `BenchmarkQuerySet_MultipleFilters`: 测试多组过滤条件的性能
- `BenchmarkQuerySet_Where`: 测试直接WHERE条件的性能
- `BenchmarkQuerySet_CompleteQuery`: 测试完整查询构建的性能
- `BenchmarkQuerySet_BuildLargeQuery`: 测试大型复杂查询的性能
- `BenchmarkQuerySet_FilterExclude`: 测试过滤和排除条件组合的性能
