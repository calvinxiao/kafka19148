package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// PartitionAssignment represents the structure of a partition assignment
type PartitionAssignment struct {
	Version    int         `json:"version"`
	Partitions []Partition `json:"partitions"`
}

// Partition represents a single partition in the assignment
type Partition struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Replicas  []int  `json:"replicas"`
}

func main() {
	var (
		file   = flag.String("file", "", "Input partition assignment file")
		from   = flag.Int("from", -1, "Node ID to move partitions away from")
		to     = flag.Int("to", -1, "Node ID to move partitions to")
		stdout = flag.Bool("stdout", false, "Print three files content to stdout")
	)
	flag.Parse()

	// Validate required parameters
	if *file == "" {
		fmt.Fprintf(os.Stderr, "Error: --file parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if *from == -1 {
		fmt.Fprintf(os.Stderr, "Error: --from parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if *to == -1 {
		fmt.Fprintf(os.Stderr, "Error: --to parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Read and parse input file
	data, err := ioutil.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var assignment PartitionAssignment
	if err := json.Unmarshal(data, &assignment); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate the three reassignment plans
	plan1, plan2, plan3 := generateReassignmentPlans(assignment, *from, *to)

	// Output the plans
	if *stdout {
		printPlansToStdout(plan1, plan2, plan3)
	} else {
		savePlansToFiles(*file, plan1, plan2, plan3)
	}
}

// generateReassignmentPlans creates three reassignment plans to avoid unclean leader election
func generateReassignmentPlans(assignment PartitionAssignment, from, to int) (PartitionAssignment, PartitionAssignment, PartitionAssignment) {
	plan1 := PartitionAssignment{Version: assignment.Version}
	plan2 := PartitionAssignment{Version: assignment.Version}
	plan3 := PartitionAssignment{Version: assignment.Version}

	for _, partition := range assignment.Partitions {
		// Check if partition needs to be modified
		if !containsInt(partition.Replicas, from) || containsInt(partition.Replicas, to) {
			// Skip this partition if 'from' is not in replicas or 'to' is already in replicas
			continue
		}

		// Plan 1: Add new replica to the end
		plan1Replicas := make([]int, len(partition.Replicas))
		copy(plan1Replicas, partition.Replicas)
		plan1Replicas = append(plan1Replicas, to)

		plan1.Partitions = append(plan1.Partitions, Partition{
			Topic:     partition.Topic,
			Partition: partition.Partition,
			Replicas:  plan1Replicas,
		})

		// Plan 2: Move new replica to the front (elect new leader)
		plan2Replicas := make([]int, len(plan1Replicas))
		plan2Replicas[0] = to
		idx := 1
		for _, replica := range partition.Replicas {
			plan2Replicas[idx] = replica
			idx++
		}

		plan2.Partitions = append(plan2.Partitions, Partition{
			Topic:     partition.Topic,
			Partition: partition.Partition,
			Replicas:  plan2Replicas,
		})

		// Plan 3: Remove old replica
		plan3Replicas := make([]int, 0, len(plan2Replicas)-1)
		for _, replica := range plan2Replicas {
			if replica != from {
				plan3Replicas = append(plan3Replicas, replica)
			}
		}

		plan3.Partitions = append(plan3.Partitions, Partition{
			Topic:     partition.Topic,
			Partition: partition.Partition,
			Replicas:  plan3Replicas,
		})
	}

	return plan1, plan2, plan3
}

// containsInt checks if a slice contains a specific integer
func containsInt(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// printPlansToStdout prints all three plans to stdout
func printPlansToStdout(plan1, plan2, plan3 PartitionAssignment) {
	fmt.Println("=== Plan 1: Add new replica ===")
	printPlan(plan1)
	fmt.Println("\n=== Plan 2: Elect new leader ===")
	printPlan(plan2)
	fmt.Println("\n=== Plan 3: Remove old replica ===")
	printPlan(plan3)
}

// printPlan prints a single plan to stdout
func printPlan(plan PartitionAssignment) {
	data, err := json.MarshalIndent(plan, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// savePlansToFiles saves the three plans to separate files
func savePlansToFiles(inputFile string, plan1, plan2, plan3 PartitionAssignment) {
	baseFilename := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	saveFile := func(plan PartitionAssignment, suffix string) {
		filename := baseFilename + suffix + ".json"
		data, err := json.MarshalIndent(plan, "", "\t")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON for %s: %v\n", filename, err)
			return
		}

		if err := ioutil.WriteFile(filename, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", filename, err)
			return
		}

		fmt.Printf("Saved: %s\n", filename)
	}

	saveFile(plan1, "-01")
	saveFile(plan2, "-02")
	saveFile(plan3, "-03")
}
