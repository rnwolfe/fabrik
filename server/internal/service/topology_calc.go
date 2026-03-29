package service

import (
	"fmt"
	"math"
)

// TopologyHints carries optional per-tier radix overrides and leaf-count hints.
// Fields default to zero, which means "derive from primary radix" / "full fabric".
type TopologyHints struct {
	// SpineRadix is the port count of a spine switch.  When non-zero it is used
	// to derive maxLeaves (each spine port connects to exactly one leaf).
	// If zero the primary radix is used as a conservative fallback.
	SpineRadix int
	// SuperSpineRadix is the port count of a super-spine switch.
	// Used in 3/5-stage calculations.  Defaults to SpineRadix when zero.
	SuperSpineRadix int
	// LeafCount overrides the leaf count: 0 = full fabric, 1 = minimum viable,
	// any other positive value is used directly (capped at the physical maximum).
	LeafCount int
}

// TopologyPlan holds the calculated switch counts and port distribution for a Clos fabric.
type TopologyPlan struct {
	// Stages is the number of Clos stages (2, 3, or 5).
	Stages int `json:"stages"`
	// Radix is the effective leaf switch port count (may be auto-corrected).
	Radix int `json:"radix"`
	// SpineRadix is the effective spine switch port count.
	SpineRadix int `json:"spine_radix,omitempty"`
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
//
// radix is the leaf switch port count.  hints (optional) carries spine radix
// overrides and a leaf-count hint:
//   - hints.LeafCount == 0  → full fabric (populate all spine ports)
//   - hints.LeafCount == 1  → minimum viable (single leaf)
//   - hints.LeafCount > 1   → use directly, capped at physical maximum
func CalculateTopology(stages int, radix int, oversubscription float64, hints ...*TopologyHints) (*TopologyPlan, error) {
	h := &TopologyHints{}
	if len(hints) > 0 && hints[0] != nil {
		h = hints[0]
	}

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
		return calc2Stage(radix, oversubscription, h)
	case 3:
		return calc3Stage(radix, oversubscription, h)
	case 5:
		return calc5Stage(radix, oversubscription, h)
	}
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
//   - uplinks = leafRadix / (1 + oversubscription)
//   - downlinks = leafRadix - uplinks
//
// spine_count = uplinks (one uplink port per leaf connects to a distinct spine).
// max_leaves  = spineRadix (each spine port connects to exactly one leaf).
// Full fabric: leaf_count = spineRadix.
// Minimum: leaf_count = 1.
// leafRadix must be divisible by (1 + oversubscription); if not, it is snapped up.
func calc2Stage(leafRadix int, oversubscription float64, h *TopologyHints) (*TopologyPlan, error) {
	divisor := int(math.Round(1.0 + oversubscription))
	if divisor < 2 {
		divisor = 2
	}

	originalRadix := leafRadix
	note := ""

	snapped, corrected := snapRadix(leafRadix, divisor)
	if corrected {
		note = fmt.Sprintf(
			"radix %d does not divide evenly for oversubscription %.2f (divisor %d); snapped to %d",
			originalRadix, oversubscription, divisor, snapped,
		)
		leafRadix = snapped
	}

	uplinks := leafRadix / divisor
	if uplinks < 1 {
		uplinks = 1
	}
	downlinks := leafRadix - uplinks
	if downlinks < 0 {
		downlinks = 0
	}

	spineCount := uplinks // one port per leaf per spine in full-mesh

	// max_leaves is determined by the spine switch port count: each spine port
	// connects to exactly one leaf in a 2-stage Clos.
	spineRadix := h.SpineRadix
	if spineRadix <= 0 {
		spineRadix = leafRadix // conservative fallback when spine model unknown
	}
	maxLeaves := spineRadix

	leafCount := maxLeaves // default: full fabric
	if h.LeafCount == 1 {
		leafCount = 1
	} else if h.LeafCount > 1 && h.LeafCount < maxLeaves {
		leafCount = h.LeafCount
	}

	plan := &TopologyPlan{
		Stages:           2,
		Radix:            leafRadix,
		SpineRadix:       spineRadix,
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
// Port allocation per leaf:
//   - uplinks = leafRadix / (1 + oversubscription)
//   - downlinks = leafRadix - uplinks
//
// spine_count = leaf_uplinks.
// super_spine_count = spineRadix / spine_count  (spine ports split between leaf and super-spine).
// max_leaves = spineRadix - super_spine_count.
func calc3Stage(leafRadix int, oversubscription float64, h *TopologyHints) (*TopologyPlan, error) {
	divisor := int(math.Round(1.0 + oversubscription))
	if divisor < 2 {
		divisor = 2
	}

	originalRadix := leafRadix
	note := ""

	snapped, corrected := snapRadix(leafRadix, divisor)
	if corrected {
		note = fmt.Sprintf(
			"radix %d does not divide evenly for oversubscription %.2f (divisor %d); snapped to %d",
			originalRadix, oversubscription, divisor, snapped,
		)
		leafRadix = snapped
	}

	uplinks := leafRadix / divisor
	if uplinks < 1 {
		uplinks = 1
	}
	downlinks := leafRadix - uplinks

	spineCount := uplinks

	// Use spine radix to determine super-spine count and max leaves.
	spineRadix := h.SpineRadix
	if spineRadix <= 0 {
		spineRadix = leafRadix
	}

	// super_spine_count = spineRadix / spineCount so that all spine uplink ports are used.
	superSpineCount := 1
	if spineCount > 0 {
		superSpineCount = spineRadix / spineCount
	}
	if superSpineCount < 1 {
		superSpineCount = 1
	}

	// Each spine: spineRadix ports = leafCount downlinks + superSpineCount uplinks.
	maxLeaves := spineRadix - superSpineCount
	if maxLeaves < 1 {
		maxLeaves = 1
	}
	leafCount := maxLeaves // default: full fabric
	if h.LeafCount == 1 {
		leafCount = 1
	} else if h.LeafCount > 1 && h.LeafCount < maxLeaves {
		leafCount = h.LeafCount
	}

	plan := &TopologyPlan{
		Stages:          3,
		Radix:           leafRadix,
		SpineRadix:      spineRadix,
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
func calc5Stage(leafRadix int, oversubscription float64, h *TopologyHints) (*TopologyPlan, error) {
	inner, err := calc3Stage(leafRadix, oversubscription, h)
	if err != nil {
		return nil, err
	}

	agg1Count := inner.SpineCount
	agg2Count := inner.SuperSpineCount

	totalSwitches := inner.LeafCount + inner.SpineCount + agg1Count + agg2Count + inner.SuperSpineCount

	return &TopologyPlan{
		Stages:              5,
		Radix:               inner.Radix,
		SpineRadix:          inner.SpineRadix,
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
