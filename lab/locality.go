package lab

import (
	"fmt"
	"regexp"
)

type Scenario struct {
	Resident string   `json:"resident"`
	SwapCost int      `json:"swap_cost"`
	MaxBurst int      `json:"max_burst"`
	Messages []string `json:"messages"`
}

type Step struct {
	Message int    `json:"message"`
	User    string `json:"user"`
	Swapped bool   `json:"swapped"`
	Start   int    `json:"start_tick"`
	Finish  int    `json:"finish_tick"`
}

type Schedule struct {
	Policy string `json:"policy"`
	Swaps  int    `json:"swaps"`
	Ticks  int    `json:"total_ticks"`
	Steps  []Step `json:"steps"`
}

type Report struct {
	Resident  string     `json:"resident"`
	SwapCost  int        `json:"swap_cost"`
	MaxBurst  int        `json:"max_burst"`
	Schedules []Schedule `json:"schedules"`
}

var userName = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

func Compare(s Scenario) (Report, error) {
	if !userName.MatchString(s.Resident) {
		return Report{}, fmt.Errorf("resident must be a short lowercase user name")
	}
	if s.SwapCost < 0 || s.SwapCost > 100 {
		return Report{}, fmt.Errorf("swap_cost must be between 0 and 100")
	}
	if s.MaxBurst < 1 || s.MaxBurst > 20 {
		return Report{}, fmt.Errorf("max_burst must be between 1 and 20")
	}
	if len(s.Messages) < 1 || len(s.Messages) > 200 {
		return Report{}, fmt.Errorf("expected 1 to 200 messages")
	}
	for i, user := range s.Messages {
		if !userName.MatchString(user) {
			return Report{}, fmt.Errorf("message %d: invalid user", i+1)
		}
	}
	return Report{
		Resident: s.Resident, SwapCost: s.SwapCost, MaxBurst: s.MaxBurst,
		Schedules: []Schedule{schedule(s, false), schedule(s, true)},
	}, nil
}

func schedule(s Scenario, locality bool) Schedule {
	result := Schedule{Policy: "fifo", Steps: make([]Step, 0, len(s.Messages))}
	if locality {
		result.Policy = "bounded locality"
	}
	pending := make([]Step, len(s.Messages))
	for i, user := range s.Messages {
		pending[i] = Step{Message: i + 1, User: user}
	}
	resident, burst := s.Resident, 0
	for len(pending) > 0 {
		index := 0
		if locality {
			index = choose(pending, resident, burst, s.MaxBurst)
		}
		step := pending[index]
		pending = append(pending[:index], pending[index+1:]...)
		if step.User != resident {
			result.Swaps++
			result.Ticks += s.SwapCost
			resident, burst = step.User, 0
			step.Swapped = true
		}
		step.Start = result.Ticks
		result.Ticks++
		step.Finish = result.Ticks
		result.Steps = append(result.Steps, step)
		burst++
	}
	return result
}

// Prefer the first queued message for the resident user until the burst cap.
// At the cap, give the earliest queued other user a turn when one is waiting.
func choose(pending []Step, resident string, burst, maxBurst int) int {
	preferResident := burst < maxBurst
	for i, step := range pending {
		matchesResident := step.User == resident
		if matchesResident == preferResident {
			return i
		}
	}
	return 0
}
