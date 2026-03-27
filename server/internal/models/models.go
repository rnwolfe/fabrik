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
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	HeightU        int       `json:"height_u"`
	PowerCapacityW int       `json:"power_capacity_w"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Rack represents a physical rack within a block (or standalone).
type Rack struct {
	ID             int64     `json:"id"`
	BlockID        *int64    `json:"block_id"`
	RackTypeID     *int64    `json:"rack_type_id"`
	Name           string    `json:"name"`
	HeightU        int       `json:"height_u"`
	PowerCapacityW int       `json:"power_capacity_w"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RackSummary is a rack with computed usage metrics and device list.
type RackSummary struct {
	Rack
	UsedU       int             `json:"used_u"`
	AvailableU  int             `json:"available_u"`
	UsedWatts   int             `json:"used_watts"`
	Devices     []*DeviceSummary `json:"devices"`
	Warning     string          `json:"warning,omitempty"`
}

// DeviceSummary is a device with its model information included.
type DeviceSummary struct {
	Device
	ModelVendor string `json:"model_vendor"`
	ModelName   string `json:"model_name"`
	HeightU     int    `json:"height_u"`
	PowerWatts  int    `json:"power_watts"`
}

// PlaceDeviceResult is returned when placing or moving a device.
type PlaceDeviceResult struct {
	Device  *Device `json:"device"`
	Warning string  `json:"warning,omitempty"`
}

// DeviceRole enumerates the role of a device within the network fabric.
type DeviceRole string

const (
	DeviceRoleSpine     DeviceRole = "spine"
	DeviceRoleLeaf      DeviceRole = "leaf"
	DeviceRoleSuperSpine DeviceRole = "super_spine"
	DeviceRoleServer    DeviceRole = "server"
	DeviceRoleOther     DeviceRole = "other"
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

// DeviceModel represents a hardware platform catalog entry (e.g., Cisco 8000, Arista 7050).
type DeviceModel struct {
	ID          int64      `json:"id"`
	Vendor      string     `json:"vendor"`
	Model       string     `json:"model"`
	PortCount   int        `json:"port_count"`
	HeightU     int        `json:"height_u"`
	PowerWatts  int        `json:"power_watts"`
	Description string     `json:"description"`
	IsSeed      bool       `json:"is_seed"`
	ArchivedAt  *time.Time `json:"archived_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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
