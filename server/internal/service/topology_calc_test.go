package service_test

import (
	"strings"
	"testing"

	"github.com/rnwolfe/fabrik/server/internal/service"
)

func TestCalculateTopology_InvalidInputs(t *testing.T) {
	tests := []struct {
		name             string
		stages           int
		radix            int
		oversubscription float64
		wantErrContains  string
	}{
		{
			name:   "invalid stages",
			stages: 4, radix: 64, oversubscription: 1.0,
			wantErrContains: "stages must be 2, 3, or 5",
		},
		{
			name:   "zero radix",
			stages: 2, radix: 0, oversubscription: 1.0,
			wantErrContains: "radix must be > 0",
		},
		{
			name:   "negative radix",
			stages: 2, radix: -1, oversubscription: 1.0,
			wantErrContains: "radix must be > 0",
		},
		{
			name:   "oversubscription below 1",
			stages: 2, radix: 64, oversubscription: 0.5,
			wantErrContains: "oversubscription must be",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CalculateTopology(tc.stages, tc.radix, tc.oversubscription)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("expected error containing %q, got %q", tc.wantErrContains, err.Error())
			}
		})
	}
}

func TestCalculateTopology_2Stage(t *testing.T) {
	tests := []struct {
		name              string
		radix             int
		oversubscription  float64
		wantSpines        int
		wantLeafDownlinks int
		wantLeafUplinks   int
		wantCorrected     bool
	}{
		{
			name:              "radix=64 os=1.0 (1:1)",
			radix:             64,
			oversubscription:  1.0,
			wantSpines:        32,
			wantLeafDownlinks: 32,
			wantLeafUplinks:   32,
		},
		{
			name:              "radix=48 os=3.0 (3:1) — divides cleanly",
			radix:             48,
			oversubscription:  3.0,
			wantSpines:        12,
			wantLeafDownlinks: 36,
			wantLeafUplinks:   12,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := service.CalculateTopology(2, tc.radix, tc.oversubscription)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plan.Stages != 2 {
				t.Errorf("expected stages=2, got %d", plan.Stages)
			}
			if plan.SpineCount != tc.wantSpines {
				t.Errorf("spines: want %d, got %d", tc.wantSpines, plan.SpineCount)
			}
			if plan.LeafUplinks != tc.wantLeafUplinks {
				t.Errorf("leaf uplinks: want %d, got %d", tc.wantLeafUplinks, plan.LeafUplinks)
			}
			if plan.LeafDownlinks != tc.wantLeafDownlinks {
				t.Errorf("leaf downlinks: want %d, got %d", tc.wantLeafDownlinks, plan.LeafDownlinks)
			}
			if tc.wantCorrected && plan.RadixCorrectionNote == "" {
				t.Error("expected radix correction note, got empty")
			}
		})
	}
}

func TestCalculateTopology_2Stage_BoundaryRadix(t *testing.T) {
	// radix=1, os=1.0 — minimum valid case
	plan, err := service.CalculateTopology(2, 1, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.SpineCount < 1 {
		t.Errorf("expected at least 1 spine, got %d", plan.SpineCount)
	}
}

func TestCalculateTopology_2Stage_RadixAutoCorrection(t *testing.T) {
	// radix=65 os=2.0 with divisor=3: 65 is not divisible by 3, should snap to 66.
	plan, err := service.CalculateTopology(2, 65, 2.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.RadixCorrectionNote == "" {
		t.Error("expected radix correction note for non-divisible radix")
	}
	if plan.OriginalRadix == 0 {
		t.Error("expected OriginalRadix to be set when correction occurs")
	}
	if plan.OriginalRadix == plan.Radix {
		t.Error("expected OriginalRadix to differ from corrected Radix")
	}
	// Corrected radix should be divisible by ceil(oversubscription)=2 → divisor=3.
	if plan.Radix%3 != 0 {
		t.Errorf("corrected radix %d not divisible by 3", plan.Radix)
	}
}

func TestCalculateTopology_3Stage(t *testing.T) {
	tests := []struct {
		name             string
		radix            int
		oversubscription float64
		wantLeafUplinks  int
		wantSpineCount   int
		wantSuperSpines  int
	}{
		{
			name:             "radix=64 os=1.0",
			radix:            64,
			oversubscription: 1.0,
			// divisor = round(1+1) = 2, uplinks=32, downlinks=32, spines=32, superspines=64/32=2
			wantLeafUplinks: 32,
			wantSpineCount:  32,
			wantSuperSpines: 2,
		},
		{
			name:             "radix=48 os=2.0",
			radix:            48,
			oversubscription: 2.0,
			// divisor = round(3) = 3, uplinks=16, downlinks=32, spines=16, superspines=48/16=3
			wantLeafUplinks: 16,
			wantSpineCount:  16,
			wantSuperSpines: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := service.CalculateTopology(3, tc.radix, tc.oversubscription)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if plan.Stages != 3 {
				t.Errorf("expected stages=3, got %d", plan.Stages)
			}
			if plan.LeafUplinks != tc.wantLeafUplinks {
				t.Errorf("leaf uplinks: want %d, got %d", tc.wantLeafUplinks, plan.LeafUplinks)
			}
			if plan.SpineCount != tc.wantSpineCount {
				t.Errorf("spine count: want %d, got %d", tc.wantSpineCount, plan.SpineCount)
			}
			if plan.SuperSpineCount != tc.wantSuperSpines {
				t.Errorf("super-spine count: want %d, got %d", tc.wantSuperSpines, plan.SuperSpineCount)
			}
		})
	}
}

func TestCalculateTopology_3Stage_RadixAutoCorrection(t *testing.T) {
	// radix=65 os=1.0: divisor=round(2)=2; 65%2 != 0, should snap to 66.
	plan, err := service.CalculateTopology(3, 65, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.RadixCorrectionNote == "" {
		t.Error("expected radix correction note")
	}
	if plan.Radix%2 != 0 {
		t.Errorf("corrected radix %d not divisible by 2", plan.Radix)
	}
}

func TestCalculateTopology_5Stage(t *testing.T) {
	plan, err := service.CalculateTopology(5, 64, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Stages != 5 {
		t.Errorf("expected stages=5, got %d", plan.Stages)
	}
	if plan.Agg1Count == 0 {
		t.Error("expected non-zero Agg1Count for 5-stage")
	}
	if plan.Agg2Count == 0 {
		t.Error("expected non-zero Agg2Count for 5-stage")
	}
	// Total switches must include all 5 tiers.
	want := plan.LeafCount + plan.SpineCount + plan.Agg1Count + plan.Agg2Count + plan.SuperSpineCount
	if plan.TotalSwitches != want {
		t.Errorf("total switches: want %d, got %d", want, plan.TotalSwitches)
	}
}

func TestCalculateTopology_DerivedMetrics(t *testing.T) {
	plan, err := service.CalculateTopology(2, 64, 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.TotalHostPorts != plan.LeafCount*plan.LeafDownlinks {
		t.Errorf("TotalHostPorts should be leafCount * leafDownlinks")
	}
	if plan.TotalSwitches != plan.LeafCount+plan.SpineCount {
		t.Errorf("TotalSwitches for 2-stage should be leafCount + spineCount")
	}
}

func TestCalculateTopology_AllStageCounts(t *testing.T) {
	for _, stages := range []int{2, 3, 5} {
		t.Run("stages="+string(rune('0'+stages)), func(t *testing.T) {
			plan, err := service.CalculateTopology(stages, 64, 1.0)
			if err != nil {
				t.Fatalf("stages=%d: unexpected error: %v", stages, err)
			}
			if plan.Stages != stages {
				t.Errorf("stages=%d: got %d", stages, plan.Stages)
			}
			if plan.TotalSwitches == 0 {
				t.Error("expected non-zero TotalSwitches")
			}
		})
	}
}

func TestCalculateTopology_OversubscriptionBoundary(t *testing.T) {
	// Minimum allowed oversubscription.
	plan, err := service.CalculateTopology(2, 64, 1.0)
	if err != nil {
		t.Fatalf("os=1.0: unexpected error: %v", err)
	}
	if plan.Oversubscription != 1.0 {
		t.Errorf("expected oversubscription=1.0, got %.2f", plan.Oversubscription)
	}

	// High oversubscription.
	plan, err = service.CalculateTopology(2, 64, 4.0)
	if err != nil {
		t.Fatalf("os=4.0: unexpected error: %v", err)
	}
	if plan.LeafUplinks >= plan.LeafDownlinks {
		t.Error("with high oversubscription, uplinks should be fewer than downlinks")
	}
}
