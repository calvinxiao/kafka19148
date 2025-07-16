# Prompt for this project

When using the built in reassignment tool to move partition to a new broker,
the current KRaft version has a bug that will cause unclean leader election.
This tool can avoid such problem by splitting the reassignemnt plan into 3 ones.

Generate Golang cli tool to split kafka partition reassignment into 3 plans.

The inputs are:

1. The current partition assignment file
2. The node.id that partitions need to move away from
3. the node.id that the partitions need to move to

Example usage:

```
kafka19148 --file input.json --from 4 --to 6 --stdout
```

params explains

- file, the input file, example format are:

```
{
	"version": 1,
	"partitions": [
		{
			"topic": "unclean-leader-election-test",
			"partition": 0,
			"replicas": [6, 5]
		}
	]
}
```

- from
- to
- stdout, print three files content to stdout, otherwise save them to current dir with -01, -02 and -03 suffix

checks to be made:

- if to appears in current replicas, do not modify such partition
- if from does not appears in current replicas, do not modify such partition
