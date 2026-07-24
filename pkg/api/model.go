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

import (
	"fmt"
	"strings"

	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
)

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

type DatacenterLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DatacenterList struct {
	Id             string              `json:"id"`
	OrganizationId string              `json:"organization_id"`
	Name           string              `json:"name"`
	Location       *DatacenterLocation `json:"location"`
	Creation       string              `json:"creation"`
	Description    string              `json:"description"`
}

type DatacenterFull = DatacenterList

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

type ClusterFull struct {
	ClusterList
	HealthStatus enum.ClusterHealthStatus `json:"health_status"`
}

type NodeList struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	ClusterId      string `json:"cluster_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IpAddress      string `json:"ip_address"`
	WolMac         string `json:"wol_mac"`
	Creation       string `json:"creation"`
	Alive          bool   `json:"alive"`
	LastSeen       string `json:"last_seen"`
	Joining        bool   `json:"joining"`
}
type NodeDetail struct {
	NodeList
	SshKey     string `json:"ssh_key"`
	ClientCert string `json:"client_cert"`
}

type NodePciDevice struct {
	Slot                 string `json:"slot"`
	Class                string `json:"class"`
	Vendor               string `json:"vendor"`
	Device               string `json:"device"`
	Revision             string `json:"revision"`
	ProgrammingInterface string `json:"prog_if"`
	IOMMUGroup           string `json:"iommu_group"`
	MarkedForPassthrough bool   `json:"marked_for_passthrough"`
}

type NodeHardwareCpu struct {
	Manufacturer string `json:"manufacturer"`
	Brand        string `json:"brand"`
	Speed        struct {
		CurrentGHz float64 `json:"current"`
	} `json:"speed"`
	Governor string `json:"governor"`
	Cores    struct {
		Physical    int `json:"physical"`
		Performance int `json:"performance"`
		Efficiency  int `json:"efficiency"`
	} `json:"cores"`
	Processors            int      `json:"processors"`
	Sockets               int      `json:"sockets"`
	Flags                 []string `json:"flags"`
	VirtualizationSupport bool     `json:"virtualization"`
}

type NodeHardwareMemory struct {
	Bank  int  `json:"bank"`
	Empty bool `json:"empty"`
	Data  struct {
		Size    int    `json:"size"`
		Type    string `json:"type"`
		ECC     bool   `json:"ecc"`
		Voltage struct {
			Current float64 `json:"current"`
			Min     float64 `json:"min"`
			Max     float64 `json:"max"`
		} `json:"voltage"`
	} `json:"data"`
}

type NodeHardwareDisk struct {
	Device      string  `json:"device"`
	Name        string  `json:"name"`
	SizeGB      float64 `json:"size"`
	Serial      string  `json:"serial"`
	Interface   string  `json:"interface"`
	SmartStatus string  `json:"smart_status"`
	Vendor      string  `json:"vendor"`
	Temperature int     `json:"temperature"`
}

type NodeHardwareUsb struct {
	Bus       int    `json:"bus"`
	Device    int    `json:"device"`
	VendorId  string `json:"vendor_id"`
	ProductId string `json:"product_id"`
	Vendor    string `json:"vendor"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Removable bool   `json:"removable"`
	MaxPower  int    `json:"max_power"`
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

type VolumeList struct {
	Id            string                `json:"id"`
	StoragePoolId string                `json:"storage_pool_id"`
	NodeId        string                `json:"node_id"`
	Name          string                `json:"name"`
	FqName        string                `json:"fq_name"`
	Description   string                `json:"description"`
	Status        enum.VolumeStatusEnum `json:"status"`
	Size          float64               `json:"size"`
	Metadata      struct {
		Driver string `json:"driver"`
	} `json:"metadata"`
	Attached   bool   `json:"attached"`
	AttachedTo string `json:"attached_to"`
	Creation   string `json:"creation"`
}

type ImageList struct {
	Name          string                `json:"name"`
	SizeMB        int64                 `json:"size"`
	Creation      string                `json:"creation"`
	Type          enum.InstanceTypeEnum `json:"type"`
	StoragePoolId string                `json:"storage_pool_id"`
}

type NodeCPUModel struct {
	Name         string `json:"name"`
	Vendor       string `json:"vendor"`
	Experimental bool   `json:"experimental"`
}

type NodeMachineType struct {
	Name     string `json:"name"`
	MaxVcpus int    `json:"max_vcpus"`
}

type NodeVirtualizationCapability struct {
	Architecture   string            `json:"arch"`
	WordSize       int               `json:"wordsize"`
	Emulator       string            `json:"emulator"`
	KVMSupported   bool              `json:"kvm_supported"`
	Firmware       map[string]any    `json:"firmware"`
	CPUModels      []NodeCPUModel    `json:"cpu_models"`
	DefaultMachine string            `json:"default_machine"`
	Machines       []NodeMachineType `json:"machines"`
	Features       map[string]bool   `json:"features"`
}

type NetworkFirewallRule struct {
	L2 struct {
		Source       string `json:"src"`
		Destination  string `json:"dst"`
		Protocol     string `json:"protocol"`
		Vlan         int    `json:"vlan"`
		VlanPriority int    `json:"vlan_priority"`
	} `json:"l2"`

	L3 struct {
		Source      string `json:"src"`
		Destination string `json:"dst"`
		Protocol    string `json:"protocol"`
	} `json:"l3"`

	L4 struct {
		Source      int `json:"src"`
		Destination int `json:"dst"`
		IcmpType    int `json:"icmp_type"`
		IcmpCode    int `json:"icmp_code"`
	} `json:"l4"`
}

type VswitchConfigSecurity struct {
	MacTable struct {
		Maximum   int `json:"maximum"`
		AgingTime int `json:"aging_time"`
	} `json:"mac_table"`

	FirewallRules []NetworkFirewallRule `json:"firewall_rules"`
}

type VswitchConfig struct {
	Security VswitchConfigSecurity `json:"security"`
	Uplinks  []string              `json:"uplinks"`
}

type VswitchList struct {
	Id             string               `json:"id"`
	Type           enum.VswitchTypeEnum `json:"type"`
	NodeId         string               `json:"node_id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Config         VswitchConfig        `json:"config"`
	Creation       string               `json:"creation"`
	PortGroupCount int                  `json:"port_group_count"`
	UplinkCount    int                  `json:"uplink_count"`
}

type StandalonePortGroupConfig struct {
	Type enum.PortGroupTypeEnum `json:"type"`
	Data struct {
		Type         enum.PortGroupTypeEnum `json:"_type"`
		VlanId       *int                   `json:"vlan_id"`
		TrunkVlans   []int                  `json:"trunk_vlans"`
		NativeVlanId *int                   `json:"native_vlan_id"`
	} `json:"data"`
	FirewallRules []NetworkFirewallRule `json:"firewall_rules"`
	Interfaces    []struct {
		Name       string `json:"name"`
		MacAddress string `json:"mac_address"`
		IPv4       string `json:"ipv4"`
		IPv6       string `json:"ipv6"`
	} `json:"interfaces"`
}

// Show summary of StandalonePortGroupConfig
func (c StandalonePortGroupConfig) String() string {
	var res strings.Builder

	switch c.Type {
	case enum.PortGroupTypeAccess:
		fmt.Fprintf(&res, "Access VLAN %d", *c.Data.VlanId)
	case enum.PortGroupTypeTrunk:
		fmt.Fprintf(&res, "Trunk VLANs %v ", c.Data.TrunkVlans)
		if c.Data.NativeVlanId != nil {
			fmt.Fprintf(&res, "(native VLAN %d)", *c.Data.NativeVlanId)
		} else {
			fmt.Fprintf(&res, "(no native VLAN)")
		}
	case enum.PortGroupTypeUntagged:
		res.WriteString("Untagged (no VLAN)")
	default:
		res.WriteString("Unknown")
	}

	return res.String()
}

type StandalonePortGroupList struct {
	VswitchId   string                    `json:"vswitch_id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Config      StandalonePortGroupConfig `json:"config"`
	Creation    string                    `json:"creation"`
}
