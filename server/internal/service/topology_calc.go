package service

import (
	"fmt"
	"math"
)

// TopologyPlan holds the calculated switch counts and port distribution for a Clos fabric.
type TopologyPlan struct {
	// Stages is the number of Clos stages (2, 3, or 5).
	Stages int `json:"stages"`
	// Radix is the effective switch port count (may be auto-corrected).
	Radix int `json:"radix"`
	// OriginalRadix is the radix value before any auto-correction (0 if no correction was made).
	OriginalRadix int `json:"original_radix,omitempty"`
	// RadixCorrectionNote explains why the radix was auto-corrected (empty if no correction).
	RadixCorrectionNote string `json:"radix_correction_note,omitempty"`
	// Oversubscription is the oversubscription ratio (downlinks:uplinks per leaf).
	Oversubscription float64 `json:"oversubscription"`

	// Per-stage switch counts.
	LeafCount       int `json:"leaf_count"`
	SpineCount      int `json:"spine_count"`
	SuperSpineCount int `json:"super_spine_count,omitempty"`
	// Aggregation layer counts for 5-stage Clos.
	Agg1Count int `json:"agg1_count,omitempty"`
	Agg2Count int `json:"agg2_count,omitempty"`

	// Port distribution per leaf.
	LeafDownlinks int `json:"leaf_downlinks"`
	LeafUplinks   int `json:"leaf_uplinks"`

	// Derived metrics.
	TotalSwitches  int `json:"total_switches"`
	TotalHostPorts int `json:"total_host_ports"`
}

// CalculateTopology computes the Clos topology for the given parameters.
// It returns a TopologyPlan or an error if the parameters are invalid.
func CalculateTopology(stages int, radix int, oversubscription float64) (*TopologyPlan, error) {
	if stages != 2 && stages != 3 && stages != 5 {
		return nil, fmt.Errorf("stages must be 2, 3, or 5, got %d", stages)
	}
	if radix <= 0 {
		return nil, fmt.Errorf("radix must be > 0, got %d", radix)
	}
	if oversubscription < 1.0 {
		return nil, fmt.Errorf("oversubscription must be >= 1.0, got %.2f", oversubscription)
	}

	switch stages {
	case 2:
		return calc2Stage(radix, oversubscription)
	case 3:
		return calc3Stage(radix, oversubscription)
	case 5:
		return calc5Stage(radix, oversubscription)
	}
	// Unreachable, but satisfies the compiler.
	return nil, fmt.Errorf("unsupported stage count: %d", stages)
}

// snapRadix finds the nearest radix value >= requested that is divisible by divisor.
// Returns the snapped value and whether a correction was made.
func snapRadix(requested int, divisor int) (snapped int, corrected bool) {
	if divisor <= 1 {
		return requested, false
	}
	if requested%divisor == 0 {
		return requested, false
	}
	snapped = int(math.Ceil(float64(requested)/float64(divisor))) * divisor
	return snapped, true
}

// calc2Stage computes a 2-stage Clos (leaf-spine) topology.
//
// Port allocation per leaf (oversubscription = downlinks/uplinks):
//   - uplinks = radix / (1 + oversubscription)
//   - downlinks = radix - uplinks
//
// Spine count equals the number of uplinks per leaf (full-mesh connectivity).
// Radix must be divisible by (1 + oversubscription); if not, it is snapped up.
func calc2Stage(radix int, oversubscription float64) (*TopologyPlan, error) {
	// Compute required integer divisor: (1 + oversubscription).
	// Use rounding to convert the float divisor to an integer.
	divisor := int(math.Round(1.0 + oversubscription))
	if divisor < 2 {
		divisor = 2
	}

	originalRadix := radix
	note := ""

	// Snap radix to a multiple of divisor for clean integer port splits.
	snapped, corrected := snapRadix(radix, divisor)
	if corrected {
		note = fmt.Sprintf(
			"radix %d does not divide evenly for oversubscription %.2f (divisor %d); snapped to %d",
			originalRadix, oversubscription, divisor, snapped,
		)
		radix = snapped
	}

	uplinks := radix / divisor
	if uplinks < 1 {
		uplinks = 1
	}
	downlinks := radix - uplinks
	if downlinks < 0 {
		downlinks = 0
	}

	spineCount := uplinks // one port per leaf per spine in full-mesh
	leafCount := 1        // logical minimum; user scales via host requirements

	plan := &TopologyPlan{
		Stages:           2,
		Radix:            radix,
		Oversubscription: oversubscription,
		LeafCount:        leafCount,
		SpineCount:       spineCount,
		LeafDownlinks:    downlinks,
		LeafUplinks:      uplinks,
		TotalSwitches:    leafCount + spineCount,
		TotalHostPorts:   leafCount * downlinks,
	}
	if note != "" {
		plan.OriginalRadix = originalRadix
		plan.RadixCorrectionNote = note
	}
	return plan, nil
}

// calc3Stage computes a 3-stage Clos (leaf-spine-superspine) topology.
//
// Port allocation per leaf (oversubscription = downlinks/uplinks):
//   - uplinks = radix / (1 + oversubscription)
//   - downlinks = radix - uplinks
//
// Spine count = leaf_uplinks (one spine connects to all leaves with one port each).
// Super-spine count = radix / spine_count.
func calc3Stage(radix int, oversubscription float64) (*TopologyPlan, error) {
	divisor := int(math.Round(1.0 + oversubscription))
	if divisor < 2 {
		divisor = 2
	}

	originalRadix := radix
	note := ""

	snapped, corrected := snapRadix(radix, divisor)
	if corrected {
		note = fmt.Sprintf(
			"radix %d does not divide evenly for oversubscription %.2f (divisor %d); snapped to %d",
			originalRadix, oversubscription, divisor, snapped,
		)
		radix = snapped
	}

	uplinks := radix / divisor
	if uplinks < 1 {
		uplinks = 1
	}
	downlinks := radix - uplinks

	spineCount := uplinks

	// Super-spine count: each spine has radix ports; needs one port to each leaf uplink.
	// With spineCount spines and each spine connecting to all leaves, spines have
	// remaining radix-leafCount ports for uplinks to super-spines.
	// Simplified formula: superSpineCount = radix / spineCount.
	superSpineCount := 1
	if spineCount > 0 {
		superSpineCount = radix / spineCount
	}
	if superSpineCount < 1 {
		superSpineCount = 1
	}

	leafCount := 1

	plan := &TopologyPlan{
		Stages:          3,
		Radix:           radix,
		Oversubscription: oversubscription,
		LeafCount:       leafCount,
		SpineCount:      spineCount,
		SuperSpineCount: superSpineCount,
		LeafDownlinks:   downlinks,
		LeafUplinks:     uplinks,
		TotalSwitches:   leafCount + spineCount + superSpineCount,
		TotalHostPorts:  leafCount * downlinks,
	}
	if note != "" {
		plan.OriginalRadix = originalRadix
		plan.RadixCorrectionNote = note
	}
	return plan, nil
}

// calc5Stage extends 3-stage with 2 additional aggregation layers.
// Layer order: leaf → spine → agg1 → agg2 → super-spine.
func calc5Stage(radix int, oversubscription float64) (*TopologyPlan, error) {
	inner, err := calc3Stage(radix, oversubscription)
	if err != nil {
		return nil, err
	}

	// Agg tiers mirror the spine and super-spine tiers.
	agg1Count := inner.SpineCount
	agg2Count := inner.SuperSpineCount

	totalSwitches := inner.LeafCount + inner.SpineCount + agg1Count + agg2Count + inner.SuperSpineCount

	return &TopologyPlan{
		Stages:              5,
		Radix:               inner.Radix,
		OriginalRadix:       inner.OriginalRadix,
		RadixCorrectionNote: inner.RadixCorrectionNote,
		Oversubscription:    oversubscription,
		LeafCount:           inner.LeafCount,
		SpineCount:          inner.SpineCount,
		SuperSpineCount:     inner.SuperSpineCount,
		Agg1Count:           agg1Count,
		Agg2Count:           agg2Count,
		LeafDownlinks:       inner.LeafDownlinks,
		LeafUplinks:         inner.LeafUplinks,
		TotalSwitches:       totalSwitches,
		TotalHostPorts:      inner.TotalHostPorts,
	}, nil
}
