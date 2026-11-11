package lab

import (
	"reflect"
	"testing"
)

func TestFIFOAndLocality(t *testing.T) {
	s := Scenario{Resident: "b", SwapCost: 3, MaxBurst: 2, Messages: []string{"a", "b", "a", "b"}}
	before := append([]string(nil), s.Messages...)
	got, err := Compare(s)
	if err != nil {
		t.Fatal(err)
	}
	want := []Schedule{
		{Policy: "fifo", Swaps: 4, Ticks: 16, Steps: []Step{
			{1, "a", true, 3, 4}, {2, "b", true, 7, 8},
			{3, "a", true, 11, 12}, {4, "b", true, 15, 16},
		}},
		{Policy: "bounded locality", Swaps: 1, Ticks: 7, Steps: []Step{
			{2, "b", false, 0, 1}, {4, "b", false, 1, 2},
			{1, "a", true, 5, 6}, {3, "a", false, 6, 7},
		}},
	}
	if !reflect.DeepEqual(got.Schedules, want) {
		t.Fatalf("schedules = %+v, want %+v", got.Schedules, want)
	}
	if !reflect.DeepEqual(before, s.Messages) {
		t.Fatal("Compare changed the input backlog")
	}
}

func TestBurstCapYieldsAndPreservesUserOrder(t *testing.T) {
	s := Scenario{Resident: "a", SwapCost: 2, MaxBurst: 2,
		Messages: []string{"a", "a", "a", "b", "c", "b", "a"}}
	got, err := Compare(s)
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []int{1, 2, 4, 6, 3, 7, 5}
	lastForUser := map[string]int{}
	seen := map[int]bool{}
	for i, step := range got.Schedules[1].Steps {
		if step.Message != wantIDs[i] {
			t.Errorf("position %d has message %d, want %d", i, step.Message, wantIDs[i])
		}
		if step.Message <= lastForUser[step.User] || seen[step.Message] {
			t.Fatalf("order or duplicate violation at %+v", step)
		}
		lastForUser[step.User] = step.Message
		seen[step.Message] = true
	}
	if len(seen) != len(s.Messages) {
		t.Fatal("messages lost")
	}
}

func TestSingleUserAndZeroCost(t *testing.T) {
	for _, s := range []Scenario{
		{Resident: "a", MaxBurst: 1, Messages: []string{"a", "a", "a"}},
		{Resident: "b", MaxBurst: 1, Messages: []string{"a", "b", "a"}},
	} {
		got, err := Compare(s)
		if err != nil {
			t.Fatal(err)
		}
		for _, schedule := range got.Schedules {
			if schedule.Ticks != len(s.Messages) {
				t.Fatalf("zero cost schedule = %+v", schedule)
			}
		}
	}
}

func TestValidation(t *testing.T) {
	valid := Scenario{Resident: "a", SwapCost: 1, MaxBurst: 1, Messages: []string{"a"}}
	cases := []Scenario{valid, valid, valid, valid, valid, valid, valid}
	cases[0].Resident = ""
	cases[1].SwapCost = -1
	cases[2].SwapCost = 101
	cases[3].MaxBurst = 0
	cases[4].Messages = nil
	cases[5].Messages = []string{"<script>"}
	cases[6].Messages = make([]string, 201)
	for i, s := range cases {
		if got, err := Compare(s); err == nil || len(got.Schedules) != 0 {
			t.Errorf("case %d = %+v, %v, want error and empty report", i, got, err)
		}
	}
}
