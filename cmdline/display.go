package cmdline

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/crackcomm/go-clitable"
)

func display(concurrency string, iterations int, interval int, stop int, concurrencyStepTime int, samples <-chan *experiment.Sample) {
	for s := range samples {
		fmt.Print("\033[2J\033[;H")
		fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
		fmt.Printf("Test underway. Concurrency: \x1b[36m%v\x1b[0m  Concurrency:TimeBetwenSteps: \x1b[36m%v\x1b[0m Workload iterations: \x1b[36m%v\x1b[0m  Interval: \x1b[36m%v\x1b[0m  Stop: \x1b[36m%v\x1b[0m\n",
			concurrency, concurrencyStepTime, iterations, interval, stop)
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

		fmt.Printf("\x1b[36mTotal iterations\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, totalIterations(iterations, interval, stop), 25), s.Total, totalIterations(iterations, interval, stop))

		fmt.Println()
		fmt.Printf("\x1b[1mLatest iteration\x1b[0m:  \x1b[36m%v\x1b[0m\n", s.LastResult)
		fmt.Printf("\x1b[1mWorst iteration\x1b[0m:   \x1b[36m%v\x1b[0m\n", s.WorstResult)
		fmt.Printf("\x1b[1mAverage iteration\x1b[0m: \x1b[36m%v\x1b[0m\n", s.Average)
		fmt.Printf("\x1b[1m95th Percentile\x1b[0m:   \x1b[36m%v\x1b[0m\n", s.NinetyfifthPercentile)
		fmt.Printf("\x1b[1mTotal time\x1b[0m:        \x1b[36m%v\x1b[0m\n", s.TotalTime)
		fmt.Printf("\x1b[1mWall time\x1b[0m:         \x1b[36m%v\x1b[0m\n", s.WallTime)
		fmt.Printf("\x1b[1mRunning Workers\x1b[0m:   \x1b[36m%v\x1b[0m\n", s.TotalWorkers)
		fmt.Println()
		fmt.Println("\x1b[32;1mCommands Issued:\x1b[0m")
		fmt.Println()
		for key, command := range s.Commands {
			fmt.Printf("\x1b[1m%v\x1b[0m:\n", key)
			fmt.Printf("\x1b[1m\tCount\x1b[0m:                 \x1b[36m%v\x1b[0m\n", command.Count)
			fmt.Printf("\x1b[1m\tAverage\x1b[0m:               \x1b[36m%v\x1b[0m\n", command.Average)
			fmt.Printf("\x1b[1m\tLast time\x1b[0m:             \x1b[36m%v\x1b[0m\n", command.LastTime)
			fmt.Printf("\x1b[1m\tWorst time\x1b[0m:            \x1b[36m%v\x1b[0m\n", command.WorstTime)
			fmt.Printf("\x1b[1m\tTotal time\x1b[0m:            \x1b[36m%v\x1b[0m\n", command.TotalTime)
			fmt.Printf("\x1b[1m\tPer second throughput\x1b[0m: \x1b[36m%v\x1b[0m\n", command.Throughput)
		}
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
		if s.TotalErrors > 0 {
			fmt.Printf("\nTotal errors: %d\n", s.TotalErrors)
			fmt.Printf("Last error: %v\n", s.LastError)
		}
		fmt.Println()
		fmt.Println("Type q <Enter> (or ctrl-c) to exit")
	}
}

func totalIterations(iterations int, interval int, stopTime int) int64 {
	var totalIterations int

	if stopTime > 0 && interval > 0 {
		totalIterations = ((stopTime / interval) + 1) * iterations
	} else {
		totalIterations = iterations
	}

	return int64(totalIterations)
}

func bar(n int64, total int64, size int) (bar string) {
	if n == 0 {
		n = 1
	}
	progress := int64(size) / (total / n)
	return "╞" + strings.Repeat("═", int(progress)) + strings.Repeat("┄", size-int(progress)) + "╡"
}

func display_table(concurrency string, iterations int, interval int, stop int, concurrencyStepTime int, samples <-chan *experiment.Sample) {
	lastErrors := make(map[string]int)
	totalError := 0
	for s := range samples {
		fmt.Print("\033[2J\033[;H")
		fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
		fmt.Printf("Test underway. Concurrency: \x1b[36m%v\x1b[0m  Concurrency:TimeBetwenSteps: \x1b[36m%v\x1b[0m Workload iterations: \x1b[36m%v\x1b[0m  Interval: \x1b[36m%v\x1b[0m  Stop: \x1b[36m%v\x1b[0m\n",
			concurrency, concurrencyStepTime, iterations, interval, stop)
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

		fmt.Printf("\x1b[36mTotal iterations\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, totalIterations(iterations, interval, stop), 25), s.Total, totalIterations(iterations, interval, stop))

		fmt.Println()
		table := New([]string{"Latest iteration", "Worst iteration", "Average iteration", "Average iteration", "95th Percentile", "Total time", "Wall time", "Running Workers"})
		table.AddRow(map[string]interface{}{
			"Latest iteration":  s.LastResult,
			"Worst iteration":   s.WorstResult,
			"Average iteration": s.Average,
			"95th Percentile":   s.NinetyfifthPercentile,
			"Total time":        s.TotalTime,
			"Wall time":         s.WallTime,
			"Running Workers":   s.TotalWorkers,
		})
		table.Markdown = true
		table.Print()
		fmt.Println()
		fmt.Println("\x1b[32;1mCommands Issued:\x1b[0m")
		fmt.Println()
		tableCmd := New([]string{"Key", "Count", "Average", "Last time", "Worst time", "Total time", "Per second throughput"})
		for key, command := range s.Commands {
			tableCmd.AddRow(map[string]interface{}{
				"Key":                   key,
				"Count":                 command.Count,
				"Average":               command.Average,
				"Last time":             command.LastTime,
				"Worst time":            command.WorstTime,
				"Total time":            command.TotalTime,
				"Per second throughput": command.Throughput,
			})
		}
		tableCmd.Markdown = true
		tableCmd.Print()
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
		if s.TotalErrors > totalError {
			totalError = s.TotalErrors
			if _, ok := lastErrors[s.LastError]; ok {
				lastErrors[s.LastError] += 1
			} else {
				lastErrors[s.LastError] = 1
			}
		}
		tableError := New([]string{"Error desc", "Count"})
		for desc, count := range lastErrors {
			tableError.AddRow(map[string]interface{}{
				"Error desc": desc,
				"Count":      count,
			})
		}
		tableError.Markdown = true
		tableError.Print()
		fmt.Println()
		fmt.Println("Type q <Enter> (or ctrl-c) to exit")
	}
}
