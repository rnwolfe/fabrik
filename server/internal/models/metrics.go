package models

// DesignMetrics holds all computed metrics for a design, including
// fabric topology, power, capacity, and port utilization.
type DesignMetrics struct {
	DesignID int64 `json:"design_id"`

	// Summary counts.
	TotalHosts    int `json:"total_hosts"`
	TotalSwitches int `json:"total_switches"`

	// BisectionBandwidthGbps is the narrowest aggregate bandwidth tier in Gbps.
	// Zero if no fabrics with device model assignments exist.
	BisectionBandwidthGbps float64 `json:"bisection_bandwidth_gbps"`

	// Oversubscription per fabric tier.
	Fabrics []FabricMetricEntry `json:"fabrics"`

	// ChokePoint is the fabric/tier with the worst-case oversubscription ratio.
	// Nil when there are no fabrics or all ratios are 1:1.
	ChokePoint *ChokePoint `json:"choke_point,omitempty"`

	// Power summary across all racks/devices in the design.
	Power PowerMetrics `json:"power"`

	// Capacity totals (compute resources).
	Capacity ResourceCapacity `json:"capacity"`

	// Port utilization per fabric.
	PortUtilization []PortUtilizationEntry `json:"port_utilization"`

	// Empty is true when the design has no devices placed.
	Empty bool `json:"empty"`
}

// FabricMetricEntry holds per-fabric oversubscription and topology metrics.
type FabricMetricEntry struct {
	FabricID   int64  `json:"fabric_id"`
	FabricName string `json:"fabric_name"`
	Tier       string `json:"tier"`
	Stages     int    `json:"stages"`

	// LeafSpineOversubscription = leaf_downlinks / leaf_uplinks.
	LeafSpineOversubscription float64 `json:"leaf_spine_oversubscription"`

	// SpineSuperSpineOversubscription is computed for 3+ stage fabrics.
	// Zero for 2-stage fabrics.
	SpineSuperSpineOversubscription float64 `json:"spine_super_spine_oversubscription"`

	// TotalSwitches across all tiers for this fabric.
	TotalSwitches int `json:"total_switches"`

	// TotalHostPorts is the number of leaf downlink (host-facing) ports.
	TotalHostPorts int `json:"total_host_ports"`
}

// ChokePoint identifies the worst-case oversubscription tier.
type ChokePoint struct {
	FabricID   int64   `json:"fabric_id"`
	FabricName string  `json:"fabric_name"`
	Tier       string  `json:"tier"`       // e.g., "leaf→spine", "spine→super-spine"
	Ratio      float64 `json:"ratio"`
}

// PowerMetrics holds power consumption totals across the design.
type PowerMetrics struct {
	// TotalCapacityW is the sum of rack power capacity (from rack definitions).
	TotalCapacityW int `json:"total_capacity_w"`

	// TotalDrawW is the sum of device typical power draw.
	TotalDrawW int `json:"total_draw_w"`

	// UtilizationPct is TotalDrawW / TotalCapacityW * 100, clamped to [0, 100].
	// Zero when TotalCapacityW is 0.
	UtilizationPct float64 `json:"utilization_pct"`
}

// ResourceCapacity holds compute resource totals.
type ResourceCapacity struct {
	TotalVCPU      int     `json:"total_vcpu"`
	TotalRAMGB     int     `json:"total_ram_gb"`
	TotalStorageTB float64 `json:"total_storage_tb"`
	TotalGPUCount  int     `json:"total_gpu_count"`
}

// PortUtilizationEntry holds port utilization for a single fabric.
type PortUtilizationEntry struct {
	FabricID      int64  `json:"fabric_id"`
	FabricName    string `json:"fabric_name"`
	TierName      string `json:"tier_name"` // e.g., "leaf", "spine", "super-spine"
	TotalPorts    int    `json:"total_ports"`
	AllocatedPorts int   `json:"allocated_ports"`
	AvailablePorts int   `json:"available_ports"`
}
