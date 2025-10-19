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
