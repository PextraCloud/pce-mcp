// For the new api code for the creating an instance
package api

// Retrieve standalone vSwitch
import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
)

type RetrieveStandaloneVSwitchesArg struct {
	NodeId string
}

type RetrieveStandaloneVSwitchesResponse []struct {
	Id          string `json:"id"`
	Type        int    `json:"type"`
	NodeId      string `json:"node_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      struct {
		Security struct {
			MacTable struct {
				Maximum   int `json:"maximum"`
				AgingTime int `json:"aging_time"`
			} `json:"mac_table"`

			FirewallRules []struct {
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
			} `json:"firewall_rules"`
		} `json:"security"`

		Uplinks []string `json:"uplinks"`
	} `json:"config"`

	Creation       *string `json:"creation"`
	PortGroupCount string  `json:"port_group_count"`
	UplinkCount    string  `json:"uplink_count"`
}

func RetrieveStandaloneVSwitches(
	ctx context.Context,
	c *Client,
	arg *RetrieveStandaloneVSwitchesArg,
) (*RetrieveStandaloneVSwitchesResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath(
		"/v1/networks/vswitches/standalone/",
		nil,
	)

	query := url.Values{
		"node_id": []string{arg.NodeId},
	}

	var resp RetrieveStandaloneVSwitchesResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// List Volumes By Node Or StoragePool
type ListVolumesByNodeOrStoragePoolArg struct {
	NodeId        string
	StoragePoolId string
}

type ListVolumesByNodeOrStoragePoolResponse []struct {
	Id            string  `json:"id"`
	StoragePoolId string  `json:"storage_pool_id"`
	NodeId        string  `json:"node_id"`
	Name          string  `json:"name"`
	FqName        string  `json:"fq_name"`
	Description   string  `json:"description"`
	Status        int     `json:"status"`
	Size          float64 `json:"size"`
	Metadata      struct {
		Driver string `json:"driver"`
	} `json:"metadata"`
	Attached   bool    `json:"attached"`
	AttachedTo string  `json:"attached_to"`
	Creation   *string `json:"creation"`
}

func ListVolumesByNodeOrStoragePool(
	ctx context.Context,
	c *Client,
	arg *ListVolumesByNodeOrStoragePoolArg,
) (*ListVolumesByNodeOrStoragePoolResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/volumes/", nil)
	query := url.Values{"node_id": []string{arg.NodeId}}
	if arg.StoragePoolId != "" {
		query.Set("storage_pool_id", arg.StoragePoolId)
	}

	var resp ListVolumesByNodeOrStoragePoolResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
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

// Retrieve Stand alone Port Groups
type PortGroupInterfaceAddress struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

type RetrieveStandalonePortGroupsArg struct {
	NodeId string
}

type RetrieveStandalonePortGroupsResponse []struct {
	VSwitchId   string `json:"vswitch_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      struct {
		Type int `json:"type"`
		Data struct {
			Type   int `json:"_type"`
			VlanId int `json:"vlan_id"`
		} `json:"data"`
		FirewallRules []NetworkFirewallRule `json:"firewall_rules"`
		Interfaces    []struct {
			Name       string                    `json:"name"`
			MacAddress string                    `json:"mac_address"`
			IPv4       PortGroupInterfaceAddress `json:"ipv4"`
			IPv6       PortGroupInterfaceAddress `json:"ipv6"`
		} `json:"interfaces"`
	} `json:"config"`
	Creation *string `json:"creation"`
}

func RetrieveStandalonePortGroups(
	ctx context.Context,
	c *Client,
	arg *RetrieveStandalonePortGroupsArg,
) (*RetrieveStandalonePortGroupsResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath("/v1/networks/vswitches/standalone/port_groups", nil)
	query := url.Values{"node_id": []string{arg.NodeId}}

	var resp RetrieveStandalonePortGroupsResponse
	if apiErr := c.Get(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// RetrieveNodeVirtualizationCapabilities

type RetrieveNodeVirtualizationCapabilitiesArg struct {
	NodeId string
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

type RetrieveNodeVirtualizationCapabilitiesResponse map[string]NodeVirtualizationCapability

func RetrieveNodeVirtualizationCapabilities(
	ctx context.Context,
	c *Client,
	arg *RetrieveNodeVirtualizationCapabilitiesArg,
) (*RetrieveNodeVirtualizationCapabilitiesResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath(
		"/v1/nodes/{node_id}/capabilities",
		map[string]string{"node_id": arg.NodeId},
	)

	var resp RetrieveNodeVirtualizationCapabilitiesResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// DeployInstance
type DeployInstanceCPU struct {
	Sockets  int    `json:"sockets"`
	Cores    int    `json:"cores"`
	Threads  int    `json:"threads"`
	Affinity string `json:"affinity,omitempty"`
}

type DeployInstanceIPAddress struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type DeployInstanceBandwidthRate struct {
	Average string `json:"average,omitempty"`
	Burst   string `json:"burst,omitempty"`
}

type DeployInstanceBandwidth struct {
	Inbound  DeployInstanceBandwidthRate `json:"inbound"`
	Outbound DeployInstanceBandwidthRate `json:"outbound"`
}

type DeployInstanceNetwork struct {
	Mac           string                   `json:"mac,omitempty"`
	VSwitchId     string                   `json:"vswitch_id"`
	PortGroupName string                   `json:"port_group_name"`
	Bandwidth     *DeployInstanceBandwidth `json:"bandwidth,omitempty"`
	Model         string                   `json:"model,omitempty"`
	IPv4          *DeployInstanceIPAddress `json:"ipv4,omitempty"`
	IPv6          *DeployInstanceIPAddress `json:"ipv6,omitempty"`
}

type DeployQEMUSecureBoot struct {
	Enabled            bool `json:"enabled"`
	EnrollStandardKeys bool `json:"enroll_standard_keys"`
}

type DeployQEMUNewVolume struct {
	Backup        bool    `json:"backup"`
	ReadOnly      bool    `json:"readonly"`
	Bus           string  `json:"bus"`
	Device        string  `json:"dev"`
	Driver        string  `json:"driver"`
	StoragePoolId string  `json:"storage_pool_id"`
	Size          float64 `json:"size"`
}

type DeployQEMUExistingVolume struct {
	Backup   bool   `json:"backup"`
	ReadOnly bool   `json:"readonly"`
	Bus      string `json:"bus"`
	Device   string `json:"dev"`
	Driver   string `json:"driver"`
	VolumeId string `json:"volume_id"`
}

type DeployQEMUInstanceMetadata struct {
	Type            string                     `json:"_type"`
	CPUModel        string                     `json:"cpu_model"`
	MachineType     string                     `json:"machine_type"`
	Firmware        string                     `json:"firmware"`
	SecureBoot      DeployQEMUSecureBoot       `json:"secure_boot"`
	NewVolumes      []DeployQEMUNewVolume      `json:"new_volumes"`
	ExistingVolumes []DeployQEMUExistingVolume `json:"existing_volumes"`
}

type DeployLXCInstanceMetadata struct {
	Type string `json:"_type"`
	Init string `json:"init,omitempty"`
}

type DeployInstanceV2Arg struct {
	Type         enum.InstanceTypeEnum
	NodeId       string
	Name         string
	Description  string
	Architecture string
	Image        string
	CPU          DeployInstanceCPU
	Memory       int
	IMDSEnabled  bool
	Networks     []DeployInstanceNetwork
	LXCMetadata  *DeployLXCInstanceMetadata
	QEMUMetadata *DeployQEMUInstanceMetadata
	Autostart    bool
}

type DeployInstanceV2Response struct {
	TaskId string `json:"task_id"`
}

type deployInstanceV2Payload struct {
	Destination struct {
		NodeId string `json:"node_id"`
	} `json:"destination"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description,omitempty"`
	Architecture string                  `json:"architecture"`
	Image        string                  `json:"image,omitempty"`
	Type         enum.InstanceTypeEnum   `json:"type"`
	CPU          DeployInstanceCPU       `json:"cpu"`
	Memory       int                     `json:"memory"`
	IMDSEnabled  bool                    `json:"imds_enabled"`
	Networks     []DeployInstanceNetwork `json:"networks"`
	Metadata     any                     `json:"metadata"`
	Autostart    bool                    `json:"autostart"`
}

func DeployInstanceV2(
	ctx context.Context,
	c *Client,
	arg *DeployInstanceV2Arg,
) (*DeployInstanceV2Response, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "deployment arguments are required")
	}
	if arg.NodeId == "" || arg.Name == "" || arg.Architecture == "" {
		return nil, NewAPIError(400, "node_id, name, and architecture are required")
	}
	if arg.CPU.Sockets <= 0 || arg.CPU.Cores <= 0 || arg.CPU.Threads <= 0 {
		return nil, NewAPIError(400, "CPU sockets, cores, and threads must be greater than zero")
	}
	if arg.Memory <= 0 {
		return nil, NewAPIError(400, "memory must be greater than zero")
	}

	var metadata any
	switch arg.Type {
	case enum.InstanceTypeEnumLXC:
		if arg.Image == "" {
			return nil, NewAPIError(400, "image is required for LXC instances")
		}

		lxcMetadata := DeployLXCInstanceMetadata{}
		if arg.LXCMetadata != nil {
			lxcMetadata = *arg.LXCMetadata
		}
		lxcMetadata.Type = "lxc"
		metadata = lxcMetadata

	case enum.InstanceTypeEnumQEMU:
		if arg.QEMUMetadata == nil {
			return nil, NewAPIError(400, "qemu_metadata is required for QEMU instances")
		}

		qemuMetadata := *arg.QEMUMetadata
		if qemuMetadata.CPUModel == "" || qemuMetadata.MachineType == "" || qemuMetadata.Firmware == "" {
			return nil, NewAPIError(400, "cpu_model, machine_type, and firmware are required for QEMU instances")
		}
		if qemuMetadata.Firmware != "bios" && qemuMetadata.Firmware != "efi" {
			return nil, NewAPIError(400, "firmware must be either bios or efi")
		}

		qemuMetadata.Type = "qemu"
		if qemuMetadata.NewVolumes == nil {
			qemuMetadata.NewVolumes = []DeployQEMUNewVolume{}
		}
		if qemuMetadata.ExistingVolumes == nil {
			qemuMetadata.ExistingVolumes = []DeployQEMUExistingVolume{}
		}
		metadata = qemuMetadata

	default:
		return nil, NewAPIError(400, "only LXC and QEMU instance types are supported")
	}

	networks := arg.Networks
	if networks == nil {
		networks = []DeployInstanceNetwork{}
	}

	payload := deployInstanceV2Payload{
		Name:         arg.Name,
		Description:  arg.Description,
		Architecture: arg.Architecture,
		Image:        arg.Image,
		Type:         arg.Type,
		CPU:          arg.CPU,
		Memory:       arg.Memory,
		IMDSEnabled:  arg.IMDSEnabled,
		Networks:     networks,
		Metadata:     metadata,
		Autostart:    arg.Autostart,
	}
	payload.Destination.NodeId = arg.NodeId

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode request payload")
	}

	path := c.ExpandPath("/v2/instances/", nil)

	var resp DeployInstanceV2Response
	if apiErr := c.Post(ctx, path, nil, bytes.NewReader(data), &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}
