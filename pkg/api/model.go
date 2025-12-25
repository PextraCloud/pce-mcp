/*
Copyright 2025 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import "github.com/PextraCloud/pce-mcp/pkg/api/enum"

type OrganizationDetail struct {
	Organization struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Creation    string `json:"creation"`
		Description string `json:"description"`
	} `json:"organization"`
	Datacenters []DatacenterList `json:"datacenters"`
	Clusters    []ClusterList    `json:"clusters"`
	Nodes       []NodeList       `json:"nodes"`
}

type UserList struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	Username       string `json:"username"`
	Enabled        bool   `json:"enabled"`
	Locked         bool   `json:"locked"`
	Expiry         string `json:"expiry"`
	Expired        bool   `json:"expired"`
	MfaEnabled     bool   `json:"mfa_enabled"`
	IsRoot         bool   `json:"is_root"`
	LinuxUser      string `json:"linux_user"`
	Description    string `json:"description"`
	Creation       string `json:"creation"`
}

type DatacenterList struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	Name           string `json:"name"`
	Location       struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Creation    string `json:"creation"`
	Description string `json:"description"`
}

type ClusterList struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	DatacenterId   string `json:"datacenter_id"`
	Name           string `json:"name"`
	Creation       string `json:"creation"`
	Description    string `json:"description"`
	NodeCount      int    `json:"node_count"`
	FaultTolerance int    `json:"fault_tolerance"`
	Standalone     bool   `json:"standalone"`
	LeaderId       string `json:"leader_id"`
	HasLeader      bool   `json:"has_leader"`
}

type InstanceList struct {
	Id     string `json:"id"`
	NodeId string `json:"node_id"`
	Type   int    `json:"type"`
	Name   string `json:"name"`
	Cpu    struct {
		Sockets int `json:"sockets"`
		Cores   int `json:"cores"`
		Threads int `json:"threads"`
	} `json:"cpu"`
	Vcpus     int    `json:"vcpus"`
	Memory    int    `json:"memory"`
	Creation  string `json:"creation"`
	Autostart bool   `json:"autostart"`
	BootOrder int    `json:"boot_order"`
}

type StoragePoolDetail struct {
	Id            string                   `json:"id"`
	Type          enum.StoragePoolTypeEnum `json:"type"`
	Name          string                   `json:"name"`
	Initialized   bool                     `json:"initialized"`
	Available     bool                     `json:"available"`
	CanHoldImages bool                     `json:"can_hold_images"`
	Usage         struct {
		CapacityGB  float64 `json:"capacity"`
		AllocatedGB float64 `json:"allocated"`
		AvailableGB float64 `json:"available"`
		PercentUsed float64 `json:"percent_used"`
	} `json:"usage"`
	VolumeCount int `json:"volume_count"`
}

type ImageList struct {
	Name          string                `json:"name"`
	SizeMB        int64                 `json:"size"`
	Creation      string                `json:"creation"`
	Type          enum.InstanceTypeEnum `json:"type"`
	StoragePoolId string                `json:"storage_pool_id"`
}
