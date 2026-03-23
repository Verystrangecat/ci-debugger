package workflow

import (
	"fmt"
	"sort"
	"strings"
)

// ExpandMatrix returns all matrix combinations for a job as a slice of
// variable maps (name → string value). Returns nil if the job has no matrix.
func ExpandMatrix(job *Job) []map[string]string {
	if job.Strategy == nil || len(job.Strategy.Matrix) == 0 {
		return nil
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(job.Strategy.Matrix))
	for k := range job.Strategy.Matrix {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Cartesian product
	combos := []map[string]string{{}}
	for _, key := range keys {
		values := job.Strategy.Matrix[key]
		newCombos := make([]map[string]string, 0, len(combos)*len(values))
		for _, combo := range combos {
			for _, val := range values {
				newCombo := make(map[string]string, len(combo)+1)
				for k, v := range combo {
					newCombo[k] = v
				}
				newCombo[key] = fmt.Sprintf("%v", val)
				newCombos = append(newCombos, newCombo)
			}
		}
		combos = newCombos
	}

	return combos
}

// MatrixSuffix returns a short suffix like "(node-version=18, os=ubuntu)"
// used for display and job ID disambiguation.
func MatrixSuffix(combo map[string]string) string {
	if len(combo) == 0 {
		return ""
	}
	keys := make([]string, 0, len(combo))
	for k := range combo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+combo[k])
	}
	return " (" + strings.Join(parts, ", ") + ")"
}
