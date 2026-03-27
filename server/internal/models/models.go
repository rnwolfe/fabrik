// Package models defines the core domain types for fabrik.
// These types are the canonical data model and are mirrored as TypeScript interfaces
// in frontend/src/app/models/.
package models

import "time"

// Design represents a top-level datacenter network design project.
type Design struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Site represents a physical datacenter location.
type Site struct {
	ID          int64     `json:"id"`
	DesignID    int64     `json:"design_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SuperBlock represents a group of blocks within a site (e.g., a data hall or pod).
type SuperBlock struct {
	ID          int64     `json:"id"`
	SiteID      int64     `json:"site_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Block represents a logical grouping of racks within a super-block (e.g., a row or cluster).
type Block struct {
	ID           int64     `json:"id"`
	SuperBlockID int64     `json:"super_block_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RackTemplate represents a named rack type template that defines hardware specifications.
type RackTemplate struct {
	ID                  int64     `json:"id"`
	Name                string    `json:"name"`
	HeightU             int       `json:"height_u"`
	PowerCapacityW      int       `json:"power_capacity_w"`
	PowerOversubPctWarn int       `json:"power_oversub_pct_warn"`
	PowerOversubPctMax  int       `json:"power_oversub_pct_max"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Rack represents a physical rack within a block (or standalone).
type Rack struct {
	ID                  int64     `json:"id"`
	BlockID             *int64    `json:"block_id"`
	RackTypeID          *int64    `json:"rack_type_id"`
	Name                string    `json:"name"`
	HeightU             int       `json:"height_u"`
	PowerCapacityW      int       `json:"power_capacity_w"`
	PowerOversubPctWarn int       `json:"power_oversub_pct_warn"`
	PowerOversubPctMax  int       `json:"power_oversub_pct_max"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// RackSummary is a rack with computed usage metrics and device list.
type RackSummary struct {
	Rack
	UsedU              int              `json:"used_u"`
	AvailableU         int              `json:"available_u"`
	UsedWattsIdle      int              `json:"used_watts_idle"`
	UsedWattsTypical   int              `json:"used_watts_typical"`
	UsedWattsMax       int              `json:"used_watts_max"`
	Devices            []*DeviceSummary `json:"devices"`
	Warning            string           `json:"warning,omitempty"`
}

// DeviceSummary is a device with its model information included.
type DeviceSummary struct {
	Device
	ModelVendor        string  `json:"model_vendor"`
	ModelName          string  `json:"model_name"`
	ModelType          string  `json:"model_type"`
	HeightU            int     `json:"height_u"`
	PowerWattsIdle     int     `json:"power_watts_idle"`
	PowerWattsTypical  int     `json:"power_watts_typical"`
	PowerWattsMax      int     `json:"power_watts_max"`
	CPUSockets         int     `json:"cpu_sockets"`
	CoresPerSocket     int     `json:"cores_per_socket"`
	RAMGB              int     `json:"ram_gb"`
	StorageTB          float64 `json:"storage_tb"`
	GPUCount           int     `json:"gpu_count"`
}

// PlaceDeviceResult is returned when placing or moving a device.
type PlaceDeviceResult struct {
	Device  *Device `json:"device"`
	Warning string  `json:"warning,omitempty"`
}

// DeviceRole enumerates the role of a device within the network fabric.
type DeviceRole string

const (
	DeviceRoleSpine         DeviceRole = "spine"
	DeviceRoleLeaf          DeviceRole = "leaf"
	DeviceRoleSuperSpine    DeviceRole = "super_spine"
	DeviceRoleServer        DeviceRole = "server"
	DeviceRoleOther         DeviceRole = "other"
	DeviceRoleManagementToR DeviceRole = "management_tor"
	DeviceRoleManagementAgg DeviceRole = "management_agg"
)

// Device represents a physical network device installed in a rack.
type Device struct {
	ID            int64      `json:"id"`
	RackID        int64      `json:"rack_id"`
	DeviceModelID int64      `json:"device_model_id"`
	Name          string     `json:"name"`
	Role          DeviceRole `json:"role"`
	Position      int        `json:"position"` // U position in rack
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PortType enumerates the physical type of a network port.
type PortType string

const (
	PortTypeEthernet PortType = "ethernet"
	PortTypeFiber    PortType = "fiber"
	PortTypeDAC      PortType = "dac"
	PortTypeOther    PortType = "other"
)

// Port represents a physical network port on a device.
type Port struct {
	ID          int64     `json:"id"`
	DeviceID    int64     `json:"device_id"`
	Name        string    `json:"name"`
	Type        PortType  `json:"type"`
	SpeedGbps   int       `json:"speed_gbps"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeviceModelType enumerates the category of a device model.
type DeviceModelType string

const (
	DeviceModelTypeNetwork DeviceModelType = "network"
	DeviceModelTypeServer  DeviceModelType = "server"
	DeviceModelTypeStorage DeviceModelType = "storage"
	DeviceModelTypeOther   DeviceModelType = "other"
)

// DeviceModel represents a hardware platform catalog entry (e.g., Cisco 8000, Arista 7050).
type DeviceModel struct {
	ID                int64           `json:"id"`
	Vendor            string          `json:"vendor"`
	Model             string          `json:"model"`
	DeviceModelType   DeviceModelType `json:"device_model_type"`
	PortCount         int             `json:"port_count"`
	HeightU           int             `json:"height_u"`
	PowerWattsIdle    int             `json:"power_watts_idle"`
	PowerWattsTypical int             `json:"power_watts_typical"`
	PowerWattsMax     int             `json:"power_watts_max"`
	// Server resource fields (0 for non-server models)
	CPUSockets      int     `json:"cpu_sockets"`
	CoresPerSocket  int     `json:"cores_per_socket"`
	RAMGB           int     `json:"ram_gb"`
	StorageTB       float64 `json:"storage_tb"`
	GPUCount        int     `json:"gpu_count"`
	Description     string  `json:"description"`
	IsSeed          bool    `json:"is_seed"`
	ArchivedAt      *time.Time `json:"archived_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}


// CapacityLevel enumerates the hierarchy level at which capacity is requested.
type CapacityLevel string

const (
	CapacityLevelRack       CapacityLevel = "rack"
	CapacityLevelBlock      CapacityLevel = "block"
	CapacityLevelSuperBlock CapacityLevel = "superblock"
	CapacityLevelSite       CapacityLevel = "site"
	CapacityLevelDesign     CapacityLevel = "design"
)

// CapacitySummary holds aggregated power and resource totals for a hierarchy level.
type CapacitySummary struct {
	Level             CapacityLevel `json:"level"`
	ID                int64         `json:"id"`
	Name              string        `json:"name"`
	PowerWattsIdle    int           `json:"power_watts_idle"`
	PowerWattsTypical int           `json:"power_watts_typical"`
	PowerWattsMax     int           `json:"power_watts_max"`
	TotalVCPU         int           `json:"total_vcpu"`
	TotalRAMGB        int           `json:"total_ram_gb"`
	TotalStorageTB    float64       `json:"total_storage_tb"`
	TotalGPUCount     int           `json:"total_gpu_count"`
	DeviceCount       int           `json:"device_count"`
}

// NetworkPlane enumerates the network plane for block aggregation assignments.
type NetworkPlane string

const (
	// PlaneFrontEnd / NetworkPlaneFrontEnd: the front-end (leaf/spine) fabric plane.
	PlaneFrontEnd       NetworkPlane = "front_end"
	NetworkPlaneFrontEnd NetworkPlane = "front_end"

	// PlaneManagement / NetworkPlaneManagement: the management network plane.
	PlaneManagement       NetworkPlane = "management"
	NetworkPlaneManagement NetworkPlane = "management"
)

// BlockAggregation represents an aggregation switch model assigned to a block for a given plane.
// One aggregation switch is allowed per (block, plane) pair.
type BlockAggregation struct {
	ID            int64        `json:"id"`
	BlockID       int64        `json:"block_id"`
	Plane         NetworkPlane `json:"plane"`
	DeviceModelID int64        `json:"device_model_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// BlockAggregationSummary extends BlockAggregation with capacity utilization.
type BlockAggregationSummary struct {
	BlockAggregation
	TotalPorts      int    `json:"total_ports"`
	AllocatedPorts  int    `json:"allocated_ports"`
	AvailablePorts  int    `json:"available_ports"`
	Utilization     string `json:"utilization"`
	Warning         string `json:"warning,omitempty"`
}

// PortConnection represents a single port allocation between a rack and a block's agg switch.
type PortConnection struct {
	ID                  int64     `json:"id"`
	BlockAggregationID  int64     `json:"block_aggregation_id"`
	RackID              int64     `json:"rack_id"`
	AggPortIndex        int       `json:"agg_port_index"`
	LeafDeviceName      string    `json:"leaf_device_name"`
	CreatedAt           time.Time `json:"created_at"`
}

// AddRackToBlockResult is returned when a rack is added to a block.
type AddRackToBlockResult struct {
	Rack        *Rack                  `json:"rack"`
	Connections []*PortConnection      `json:"connections"`
	Warning     string                 `json:"warning,omitempty"`
}

// FabricTier enumerates whether a fabric tier is a front-end or back-end network.
type FabricTier string

const (
	FabricTierFrontEnd FabricTier = "frontend"
	FabricTierBackEnd  FabricTier = "backend"
)

// Fabric represents a Clos fabric tier (front-end or back-end) within a design.
type Fabric struct {
	ID          int64      `json:"id"`
	DesignID    int64      `json:"design_id"`
	Name        string     `json:"name"`
	Tier        FabricTier `json:"tier"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
