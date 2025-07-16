package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestGenerateReassignmentPlans_Basic(t *testing.T) {
	assignment := PartitionAssignment{
		Version: 1,
		Partitions: []Partition{
			{
				Topic:     "test-topic",
				Partition: 0,
				Replicas:  []int{4, 5},
			},
		},
	}
	plan1, plan2, plan3 := generateReassignmentPlans(assignment, 4, 6)

	wantPlan1 := PartitionAssignment{
		Version: 1,
		Partitions: []Partition{
			{
				Topic:     "test-topic",
				Partition: 0,
				Replicas:  []int{4, 5, 6},
			},
		},
	}
	wantPlan2 := PartitionAssignment{
		Version: 1,
		Partitions: []Partition{
			{
				Topic:     "test-topic",
				Partition: 0,
				Replicas:  []int{6, 4, 5},
			},
		},
	}
	wantPlan3 := PartitionAssignment{
		Version: 1,
		Partitions: []Partition{
			{
				Topic:     "test-topic",
				Partition: 0,
				Replicas:  []int{6, 5},
			},
		},
	}

	if !reflect.DeepEqual(plan1, wantPlan1) {
		t.Errorf("Plan1 mismatch: got %+v, want %+v", plan1, wantPlan1)
	}
	if !reflect.DeepEqual(plan2, wantPlan2) {
		t.Errorf("Plan2 mismatch: got %+v, want %+v", plan2, wantPlan2)
	}
	if !reflect.DeepEqual(plan3, wantPlan3) {
		t.Errorf("Plan3 mismatch: got %+v, want %+v", plan3, wantPlan3)
	}
}

func TestGenerateReassignmentPlans_FromTestFiles(t *testing.T) {
	testFiles := []string{"testdata/test-input.json", "testdata/test-edge-case.json", "testdata/test-mixed.json"}
	for _, fname := range testFiles {
		data, err := os.ReadFile(fname)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", fname, err)
		}
		var assignment PartitionAssignment
		if err := json.Unmarshal(data, &assignment); err != nil {
			t.Fatalf("Failed to parse %s: %v", fname, err)
		}
		// Use 4 as 'from', 6 as 'to' for all test files
		plan1, plan2, plan3 := generateReassignmentPlans(assignment, 4, 6)
		// Just check that output is valid JSON and partitions are as expected
		for _, plan := range []PartitionAssignment{plan1, plan2, plan3} {
			if plan.Version != assignment.Version {
				t.Errorf("Version mismatch in %s", fname)
			}
			for _, p := range plan.Partitions {
				if p.Topic == "" || p.Partition < 0 || len(p.Replicas) == 0 {
					t.Errorf("Invalid partition in %s: %+v", fname, p)
				}
			}
		}
	}
}
