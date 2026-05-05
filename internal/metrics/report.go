// Package metrics – report.go provides a human-readable text summary
// of all counters collected during a vaultenv run, written to any io.Writer.
package metrics

import (
	"fmt"
	"io"
	"sort"
)

// Report writes a sorted, human-readable summary of all counter values
// in the registry to w. Counters with a value of zero are omitted.
func Report(reg *Registry, w io.Writer) error {
	snap := reg.Snapshot()
	if len(snap) == 0 {
		return nil
	}

	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintln(w, "# vaultenv metrics")
	for _, k := range keys {
		v := snap[k]
		if v == 0 {
			continue
		}
		if _, err := fmt.Fprintf(w, "%-40s %d\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

// HistogramSummary writes p50/p95/p99 percentiles for the named histogram.
// If the histogram has no samples it is a no-op.
func HistogramSummary(reg *Registry, name string, w io.Writer) error {
	samples := reg.Histogram(name).Snapshot()
	if len(samples) == 0 {
		return nil
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	pct := func(p float64) float64 {
		idx := int(p/100*float64(len(sorted)-1)+0.5)
		return sorted[idx]
	}

	_, err := fmt.Fprintf(w, "%-40s p50=%.1fms p95=%.1fms p99=%.1fms (n=%d)\n",
		name, pct(50), pct(95), pct(99), len(sorted))
	return err
}
