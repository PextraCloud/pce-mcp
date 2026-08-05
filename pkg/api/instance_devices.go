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
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

type InstanceDeviceTypeEnum int

const (
	InstanceDeviceTypeNetworkInterface      InstanceDeviceTypeEnum = 1
	InstanceDeviceTypeStorageVolume         InstanceDeviceTypeEnum = 2
	InstanceDeviceTypeTrustedPlatformModule InstanceDeviceTypeEnum = 3
	InstanceDeviceTypeUSBDevice             InstanceDeviceTypeEnum = 4
	InstanceDeviceTypePCIDevice             InstanceDeviceTypeEnum = 5
	InstanceDeviceTypeRNGDevice             InstanceDeviceTypeEnum = 6
	InstanceDeviceTypeCDROMDrive            InstanceDeviceTypeEnum = 7
)

func (e InstanceDeviceTypeEnum) IsValid() bool {
	return e >= InstanceDeviceTypeNetworkInterface &&
		e <= InstanceDeviceTypeCDROMDrive
}

// Retrieve a specific instance.

type GetInstanceByIdArg struct {
	InstanceId string
}

type GetInstanceByIdResponse struct {
	Instance InstanceList `json:"instance"`
}

func GetInstanceById(
	ctx context.Context,
	c *Client,
	arg *GetInstanceByIdArg,
) (*GetInstanceByIdResponse, *APIError) {
	if arg == nil || arg.InstanceId == "" {
		return nil, NewAPIError(400, "instance_id is required")
	}

	path := c.ExpandPath(
		"/v1/instances/{instance_id}/",
		map[string]string{"instance_id": arg.InstanceId},
	)

	var resp GetInstanceByIdResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Common instance-device structures.

type InstanceDeviceMetadata struct {
	Type InstanceDeviceTypeEnum `json:"_type"`
	Data any                    `json:"data"`
}

type InstanceDevice struct {
	Name         string                 `json:"name"`
	InstanceId   string                 `json:"instance_id"`
	Type         InstanceDeviceTypeEnum `json:"type"`
	Metadata     json.RawMessage        `json:"metadata"`
	Creation     any                    `json:"creation"`
	Description  *string                `json:"description"`
	ResourceId   string                 `json:"resource_id"`
	BootOrder    any                    `json:"boot_order"`
	IsBootDevice bool                   `json:"is_boot_device"`
}

// Attach any supported device to an instance.

type AttachDeviceToInstanceArg struct {
	NodeId      string
	InstanceId  string
	Name        string
	Description string
	Type        InstanceDeviceTypeEnum
	Metadata    InstanceDeviceMetadata
	Live        bool
}

type AttachDeviceToInstanceResponse struct {
	Name        string `json:"name"`
	NeedsReboot bool   `json:"needs_reboot"`
}

type attachDeviceToInstancePayload struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        InstanceDeviceTypeEnum `json:"type"`
	Metadata    InstanceDeviceMetadata `json:"metadata"`
	Live        bool                   `json:"live"`
}

func AttachDeviceToInstance(
	ctx context.Context,
	c *Client,
	arg *AttachDeviceToInstanceArg,
) (*AttachDeviceToInstanceResponse, *APIError) {
	if arg == nil {
		return nil, NewAPIError(400, "device attachment arguments are required")
	}
	if arg.NodeId == "" || arg.InstanceId == "" || arg.Name == "" {
		return nil, NewAPIError(400, "node_id, instance_id, and name are required")
	}
	if !arg.Type.IsValid() {
		return nil, NewAPIError(400, "invalid device type")
	}
	if arg.Metadata.Type != arg.Type {
		return nil, NewAPIError(400, "metadata type must match device type")
	}
	if arg.Metadata.Data == nil {
		return nil, NewAPIError(400, "device metadata is required")
	}

	payload := attachDeviceToInstancePayload{
		Name:        arg.Name,
		Description: arg.Description,
		Type:        arg.Type,
		Metadata:    arg.Metadata,
		Live:        arg.Live,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode device attachment payload")
	}

	path := c.ExpandPath(
		"/v1/instances/{instance_id}/devices/",
		map[string]string{"instance_id": arg.InstanceId},
	)
	query := url.Values{"node_id": []string{arg.NodeId}}

	var resp AttachDeviceToInstanceResponse
	if apiErr := c.Post(ctx, path, query, bytes.NewReader(data), &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Retrieve every device attached to an instance.

type GetAllDevicesAttachedToInstanceArg struct {
	InstanceId string
}

type GetAllDevicesAttachedToInstanceResponse []InstanceDevice

func GetAllDevicesAttachedToInstance(
	ctx context.Context,
	c *Client,
	arg *GetAllDevicesAttachedToInstanceArg,
) (*GetAllDevicesAttachedToInstanceResponse, *APIError) {
	if arg == nil || arg.InstanceId == "" {
		return nil, NewAPIError(400, "instance_id is required")
	}

	path := c.ExpandPath(
		"/v1/instances/{instance_id}/devices/",
		map[string]string{"instance_id": arg.InstanceId},
	)

	var resp GetAllDevicesAttachedToInstanceResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Update the name or description of an attached device.

type UpdateDeviceAttachedToInstanceArg struct {
	InstanceId  string
	DeviceName  string
	NewName     string
	Description *string
}

type UpdateDeviceAttachedToInstanceResponse struct{}

type updateDeviceAttachedToInstancePayload struct {
	NewName     string  `json:"new_name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func UpdateDeviceAttachedToInstance(
	ctx context.Context,
	c *Client,
	arg *UpdateDeviceAttachedToInstanceArg,
) (*UpdateDeviceAttachedToInstanceResponse, *APIError) {
	if arg == nil || arg.InstanceId == "" || arg.DeviceName == "" {
		return nil, NewAPIError(400, "instance_id and device_name are required")
	}
	if arg.NewName == "" && arg.Description == nil {
		return nil, NewAPIError(400, "new_name or description is required")
	}

	payload := updateDeviceAttachedToInstancePayload{
		NewName:     arg.NewName,
		Description: arg.Description,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, NewAPIError(500, "failed to encode device update payload")
	}

	path := c.ExpandPath(
		"/v1/instances/{instance_id}/devices/{device_name}/",
		map[string]string{
			"instance_id": arg.InstanceId,
			"device_name": arg.DeviceName,
		},
	)

	var resp UpdateDeviceAttachedToInstanceResponse
	if apiErr := c.Patch(ctx, path, nil, bytes.NewReader(data), &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Detach a device from an instance.

type DetachDeviceFromInstanceArg struct {
	NodeId     string
	InstanceId string
	DeviceName string
	Live       bool
}

type DetachDeviceFromInstanceResponse struct {
	NeedsReboot bool `json:"needs_reboot"`
}

func DetachDeviceFromInstance(
	ctx context.Context,
	c *Client,
	arg *DetachDeviceFromInstanceArg,
) (*DetachDeviceFromInstanceResponse, *APIError) {
	if arg == nil || arg.NodeId == "" || arg.InstanceId == "" || arg.DeviceName == "" {
		return nil, NewAPIError(400, "node_id, instance_id, and device_name are required")
	}

	path := c.ExpandPath(
		"/v1/instances/{instance_id}/devices/{device_name}/",
		map[string]string{
			"instance_id": arg.InstanceId,
			"device_name": arg.DeviceName,
		},
	)
	query := url.Values{
		"node_id": []string{arg.NodeId},
		"live":    []string{strconv.FormatBool(arg.Live)},
	}

	var resp DetachDeviceFromInstanceResponse
	if apiErr := c.Delete(ctx, path, query, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Retrieve USB device information for a specific node.
type RetrieveUSBDeviceInformationForNodeArg struct {
	NodeId string
}

type RetrieveUSBDeviceInformationForNodeResponse struct {
	Usb []NodeHardwareUsb `json:"usb"`
}

func RetrieveUSBDeviceInformationForNode(
	ctx context.Context,
	c *Client,
	arg *RetrieveUSBDeviceInformationForNodeArg,
) (*RetrieveUSBDeviceInformationForNodeResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath(
		"/v1/nodes/{node_id}/hardware/usb",
		map[string]string{"node_id": arg.NodeId},
	)

	var resp RetrieveUSBDeviceInformationForNodeResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// Retrieve PCI device information for a specific node.
type RetrievePCIDeviceInformationForNodeArg struct {
	NodeId string
}

type RetrievePCIDeviceInformationForNodeResponse []NodePciDevice

func RetrievePCIDeviceInformationForNode(
	ctx context.Context,
	c *Client,
	arg *RetrievePCIDeviceInformationForNodeArg,
) (*RetrievePCIDeviceInformationForNodeResponse, *APIError) {
	if arg == nil || arg.NodeId == "" {
		return nil, NewAPIError(400, "node_id is required")
	}

	path := c.ExpandPath(
		"/v1/nodes/{node_id}/hardware/pci/",
		map[string]string{"node_id": arg.NodeId},
	)

	var resp RetrievePCIDeviceInformationForNodeResponse
	if apiErr := c.Get(ctx, path, nil, &resp); apiErr != nil {
		return nil, apiErr
	}

	return &resp, nil
}

// ****** META DATA *****

// USB-specific metadata for use with AttachDeviceToInstance.

type USBStartupPolicy string

const (
	USBStartupPolicyMandatory USBStartupPolicy = "mandatory"
	USBStartupPolicyRequisite USBStartupPolicy = "requisite"
	USBStartupPolicyOptional  USBStartupPolicy = "optional"
)

type USBPhysicalDevice struct {
	VendorId  string `json:"vendor_id"`
	ProductId string `json:"product_id"`
	Bus       string `json:"bus"`
	Device    string `json:"device"`
}

type USBDeviceMetadataData struct {
	ClusterMappingId string             `json:"cluster_mapping_id,omitempty"`
	Device           *USBPhysicalDevice `json:"device,omitempty"`
	StartupPolicy    USBStartupPolicy   `json:"startup_policy,omitempty"`
}

type PCIDeviceMetadataData struct {
	Device           string `json:"device"`
	ClusterMappingId string `json:"cluster_mapping_id,omitempty"`
	IsDisplay        bool   `json:"is_display,omitempty"`
	RAMFramebuffer   bool   `json:"ramfb,omitempty"`
}
