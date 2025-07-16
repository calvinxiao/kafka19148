# Kafka 19148 CLI Tool

A Golang CLI tool to split Kafka partition reassignment into 3 plans to avoid unclean leader election in [KAFKA-19148](https://issues.apache.org/jira/browse/KAFKA-19148).

## Install

```bash
go install github.com/calvinxiao/kafka19148@latest
```

## Building

```bash
go build -o kafka19148 main.go
```

## Usage

```bash
kafka19148 --file input.json --from 4 --to 6 [--stdout]
```

### Parameters

- `--file`: Input partition assignment file (required)
- `--from`: Node ID to move partitions away from (required)
- `--to`: Node ID to move partitions to (required)
- `--stdout`: Print three files content to stdout instead of saving to files (optional)

### Input Format

The input file should be a JSON file with the following structure:

```json
{
  "version": 1,
  "partitions": [
    {
      "topic": "unclean-leader-election-test",
      "partition": 0,
      "replicas": [4, 5]
    }
  ]
}
```

### Output

The tool generates three reassignment plans:

1. **Plan 1**: Add new replica to the existing replicas
2. **Plan 2**: Move new replica to the front (elect new leader)
3. **Plan 3**: Remove old replica

#### Example

Input: `[4, 5]` (moving from node 4 to node 6)

- Plan 1: `[4, 5, 6]` - Add new replica
- Plan 2: `[6, 4, 5]` - Elect new leader
- Plan 3: `[6, 5]` - Remove old replica

### File Output

When not using `--stdout`, the tool saves three files with suffixes:

- `input-01.json` - Plan 1
- `input-02.json` - Plan 2
- `input-03.json` - Plan 3

### Validation

The tool will skip partitions if:

- The `from` node is not in the current replicas
- The `to` node is already in the current replicas

## Examples

### Print to stdout

```bash
./kafka19148 --file test-input.json --from 4 --to 6 --stdout
```

### Save to files

```bash
./kafka19148 --file test-input.json --from 4 --to 6
```

This will create:

- `test-input-01.json`
- `test-input-02.json`
- `test-input-03.json`
