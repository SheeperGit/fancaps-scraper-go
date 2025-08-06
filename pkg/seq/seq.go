package seq

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

/*
Returns a slice of integers starting from start `start` to end `end` (inclusive)
with a (strictly positive) step `step`.
*/
func generateSequence(start, end, step int) ([]int, error) {
	if step <= 0 {
		return nil, fmt.Errorf("`step` cannot be less than zero. (%d <= 0)", step)
	} else if start > end {
		return []int{}, fmt.Errorf("`start` cannot be less than `end`. (%d > %d)", start, end)
	}

	/* Include 'end' if reachable by step. */
	count := ((end - start) / step) + 1

	nums := make([]int, count)
	for i := range count {
		nums[i] = start + i*step
	}

	return nums, nil
}

/*
Returns a unique, sorted slice of integers specified by ranges in `seqStr` up to a maximum of `max`,
and .
All ranges specified in `seqStr` are inclusive.
Valid ranges can optionally include a start index, end index, and a step.

Example usage:

	parseSequenceString("1-5", 5)	// [1, 2, 3, 4, 5]
	parseSequenceString("-5", 5)	// Same as above.
	parseSequenceString("1-", 12)	// [1, ..., 12]

	parseSequenceString("1-13:3", 100)	// [1, 4, 7, 10, 13]
	parseSequenceString("1-4:2", 100)	// [1, 3]

	parseSequenceString("1", 100)	// [1]
	parseSequenceString("1, 13, 12, 19", 100)	// [1, 12, 13, 19]

	parseSequenceString("1-3, 2-4, 5-8:2, 6-8:2", 100)	// [1, 2, 3, 4, 5, 6, 7, 8]
*/
func ParseSequenceString(seqStr string, max int, debug bool) ([]int, error) {
	rangeRe := regexp.MustCompile(`^\s*(\d*)-(\d*)(?::(\d+))?\s*$`)
	intRe := regexp.MustCompile(`^\s*(\d+)\s*$`)

	uniqNums := make(map[int]struct{})

	for seq := range strings.SplitSeq(seqStr, ",") {
		seq = strings.TrimSpace(seq)

		switch {
		/* Range parsing. */
		case strings.Contains(seq, "-"):
			match := rangeRe.FindStringSubmatch(seq)
			if match == nil {
				return []int{}, fmt.Errorf("invalid range format: %s", seq)
			}

			startStr, endStr, stepStr := match[1], match[2], match[3]

			start := 1
			if startStr != "" {
				start, _ = strconv.Atoi(startStr)
				if start < 1 {
					return []int{}, fmt.Errorf("`start` cannot be less than one. (%d < 1)", start)
				}
			}

			end := max
			if endStr != "" {
				end, _ = strconv.Atoi(endStr)
				if end > max {
					return []int{}, fmt.Errorf("`end` cannot be more than `max`. (%d > %d)", end, max)
				}
			}

			step := 1
			if stepStr != "" {
				step, _ = strconv.Atoi(stepStr)
			}

			/* Generate range from sub-sequence. */
			genSeq, err := generateSequence(start, end, step)
			if err != nil {
				return nil, fmt.Errorf("failed to generate sequence %q: %w", genSeq, err)
			}
			for _, n := range genSeq {
				uniqNums[n] = struct{}{}
			}

		/* Single number parsing. */
		case !strings.Contains(seq, ":"):
			match := intRe.FindStringSubmatch(seq)
			if match == nil {
				return []int{}, fmt.Errorf("invalid single number format: %s", seq)
			}

			n, _ := strconv.Atoi(match[1])
			uniqNums[n] = struct{}{}

		/* Unknown format error. */
		default:
			return []int{}, fmt.Errorf("unknown format: %s", seq)
		}
	}

	/* Sort sequence. */
	nums := make([]int, 0, len(uniqNums))
	for n := range uniqNums {
		nums = append(nums, n)
	}
	sort.Ints(nums)

	/* Debug: Print sequence output. */
	if debug {
		fmt.Printf("You selected episodes: ")
		for i, n := range nums {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(n)
		}
		fmt.Println()
	}

	return nums, nil
}
