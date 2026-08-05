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
package pce

import (
	"context"
	"strconv"
	"strings"

	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/PextraCloud/pce-mcp/pkg/api/enum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type attachUSBDeviceOptionsResult struct {
	Message    string                `json:"message"`
	InstanceId string                `json:"instance_id"`
	NodeId     string                `json:"node_id"`
	USBDevices []api.NodeHardwareUsb `json:"usb_devices"`
	NextStep   string                `json:"next_step"`
}

func AttachUSBDevice() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("attach_usb_device",
		mcp.WithDescription("List selectable physical USB devices for a QEMU instance, then attach one using the same tool. First call with node_id and instance_id only. Show the returned USB choices to the user. After the user selects and confirms one, call again with device_name, vendor_id, product_id, usb_bus, usb_device, and confirm=true."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Attach USB Device",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(true),
			IdempotentHint:  mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Node hosting both the QEMU instance and physical USB device"),
		),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("QEMU instance that will receive the USB device"),
		),
		mcp.WithString("device_name",
			mcp.MinLength(3),
			mcp.MaxLength(64),
			mcp.Pattern("^[a-zA-Z0-9_-]+$"),
			mcp.Description("Unique attachment name inside the instance, for example usb_keyboard"),
		),
		mcp.WithString("description",
			mcp.MaxLength(descriptionDefaultMaxLength),
			mcp.Description("Optional description of the USB attachment"),
		),
		mcp.WithString("vendor_id",
			mcp.Pattern("^[0-9A-Fa-f]{4}$"),
			mcp.Description("Four-character hexadecimal USB vendor ID from the selected USB option"),
		),
		mcp.WithString("product_id",
			mcp.Pattern("^[0-9A-Fa-f]{4}$"),
			mcp.Description("Four-character hexadecimal USB product ID from the selected USB option"),
		),
		mcp.WithNumber("usb_bus",
			mcp.Min(1),
			mcp.Description("USB bus number from the selected USB option"),
		),
		mcp.WithNumber("usb_device",
			mcp.Min(1),
			mcp.Description("USB device number from the selected USB option"),
		),
		mcp.WithString("startup_policy",
			mcp.Enum(
				string(api.USBStartupPolicyMandatory),
				string(api.USBStartupPolicyRequisite),
				string(api.USBStartupPolicyOptional),
			),
			mcp.Description("Optional behavior when the USB device is unavailable at instance startup"),
		),
		mcp.WithBoolean("live",
			mcp.DefaultBool(false),
			mcp.Description("Attempt to hot-plug the USB device while the instance is running"),
		),
		mcp.WithBoolean("confirm",
			mcp.DefaultBool(false),
			mcp.Description("Must be true only after the user confirms the node, instance, and physical USB device"),
		),
	), handleAttachUSBDevice
}

func handleAttachUSBDevice(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceId, err := requiredParam[string](req, "instance_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First internal API call: retrieve and validate the instance.
	instance, getErr := api.GetInstanceById(ctx, client, &api.GetInstanceByIdArg{
		InstanceId: instanceId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}
	if instance.Instance.NodeId != nodeId {
		return mcp.NewToolResultError("the selected instance is not hosted on the selected node"), nil
	}
	if instance.Instance.Type != int(enum.InstanceTypeEnumQEMU) {
		return mcp.NewToolResultError("physical USB devices can only be attached to QEMU instances"), nil
	}

	// Second internal API call: retrieve the physical USB choices for this node.
	usbInformation, usbErr := api.RetrieveUSBDeviceInformationForNode(ctx, client, &api.RetrieveUSBDeviceInformationForNodeArg{
		NodeId: nodeId,
	})
	if usbErr != nil {
		return mcp.NewToolResultError(usbErr.Error()), nil
	}
	if usbInformation.Usb == nil {
		usbInformation.Usb = make([]api.NodeHardwareUsb, 0)
	}

	arguments := req.GetArguments()
	selectionFields := []string{"vendor_id", "product_id", "usb_bus", "usb_device"}
	selectionFieldCount := 0
	for _, field := range selectionFields {
		if _, exists := arguments[field]; exists {
			selectionFieldCount++
		}
	}

	// With no selection, this is the read-only discovery phase of the tool.
	if selectionFieldCount == 0 {
		message := "Select a USB device from usb_devices before attaching it."
		if len(usbInformation.Usb) == 0 {
			message = "No USB devices are currently available on the selected node."
		}
		return mcp.NewToolResultJSON(&attachUSBDeviceOptionsResult{
			Message:    message,
			InstanceId: instanceId,
			NodeId:     nodeId,
			USBDevices: usbInformation.Usb,
			NextStep:   "Show these choices to the user. After selection and explicit confirmation, call attach_usb_device again with device_name, vendor_id, product_id, usb_bus, usb_device, and confirm=true.",
		})
	}
	if selectionFieldCount != len(selectionFields) {
		return mcp.NewToolResultError("vendor_id, product_id, usb_bus, and usb_device must all be provided when selecting a USB device"), nil
	}

	deviceName, err := optionalParam[string](req, "device_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if deviceName == "" {
		return mcp.NewToolResultError("device_name is required when attaching a selected USB device"), nil
	}
	vendorId, err := requiredParam[string](req, "vendor_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	productId, err := requiredParam[string](req, "product_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	usbBus, err := req.RequireInt("usb_bus")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	usbDevice, err := req.RequireInt("usb_device")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description, err := optionalParam[string](req, "description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	startupPolicy, err := optionalParam[string](req, "startup_policy")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !req.GetBool("confirm", false) {
		return mcp.NewToolResultError("USB attachment requires explicit user confirmation"), nil
	}

	found := false
	for _, usb := range usbInformation.Usb {
		if usb.Bus == usbBus &&
			usb.Device == usbDevice &&
			strings.EqualFold(usb.VendorId, vendorId) &&
			strings.EqualFold(usb.ProductId, productId) {
			found = true
			break
		}
	}
	if !found {
		return mcp.NewToolResultError("the selected USB device was not found in the target node hardware inventory"), nil
	}

	usbMetadata := api.USBDeviceMetadataData{
		Device: &api.USBPhysicalDevice{
			VendorId:  strings.ToLower(vendorId),
			ProductId: strings.ToLower(productId),
			Bus:       strconv.Itoa(usbBus),
			Device:    strconv.Itoa(usbDevice),
		},
		StartupPolicy: api.USBStartupPolicy(startupPolicy),
	}

	// Third internal API call: attach the USB device.
	response, attachErr := api.AttachDeviceToInstance(ctx, client, &api.AttachDeviceToInstanceArg{
		NodeId:      nodeId,
		InstanceId:  instanceId,
		Name:        deviceName,
		Description: description,
		Type:        api.InstanceDeviceTypeUSBDevice,
		Metadata: api.InstanceDeviceMetadata{
			Type: api.InstanceDeviceTypeUSBDevice,
			Data: usbMetadata,
		},
		Live: req.GetBool("live", false),
	})
	if attachErr != nil {
		return mcp.NewToolResultError(attachErr.Error()), nil
	}

	return mcp.NewToolResultJSON(struct {
		Message     string `json:"message"`
		Name        string `json:"name"`
		NeedsReboot bool   `json:"needs_reboot"`
	}{
		Message:     "USB device attached successfully",
		Name:        response.Name,
		NeedsReboot: response.NeedsReboot,
	})
}

type attachPCIDeviceOptionsResult struct {
	Message             string              `json:"message"`
	InstanceId          string              `json:"instance_id"`
	NodeId              string              `json:"node_id"`
	AvailablePCIDevices []api.NodePciDevice `json:"available_pci_devices"`
	NextStep            string              `json:"next_step"`
}

func AttachPCIDevice() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("attach_pci_device",
		mcp.WithDescription("List selectable physical PCI devices for a QEMU instance, then attach one using the same tool. First call with node_id and instance_id only. Show the returned PCI choices to the user. After the user selects and confirms one, call again with device_name, pci_slot, and confirm=true. Only devices marked for passthrough are returned as available."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Attach PCI Device",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(true),
			IdempotentHint:  mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Node hosting both the QEMU instance and physical PCI device"),
		),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("QEMU instance that will receive the PCI device"),
		),
		mcp.WithString("device_name",
			mcp.MinLength(3),
			mcp.MaxLength(64),
			mcp.Pattern("^[a-zA-Z0-9_-]+$"),
			mcp.Description("Unique attachment name inside the instance, for example pci_gpu"),
		),
		mcp.WithString("description",
			mcp.MaxLength(descriptionDefaultMaxLength),
			mcp.Description("Optional description of the PCI attachment"),
		),
		mcp.WithString("pci_slot",
			mcp.Pattern("^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}(\\.[0-7])?$"),
			mcp.Description("Exact PCI slot returned by the first call, for example 0000:65:00.0"),
		),
		mcp.WithBoolean("is_display",
			mcp.DefaultBool(false),
			mcp.Description("Enable the selected PCI device as an instance display device"),
		),
		mcp.WithBoolean("ramfb",
			mcp.DefaultBool(false),
			mcp.Description("Enable a RAM framebuffer; only valid when is_display is true"),
		),
		mcp.WithBoolean("live",
			mcp.DefaultBool(false),
			mcp.Description("Attempt to hot-plug the PCI device while the instance is running"),
		),
		mcp.WithBoolean("confirm",
			mcp.DefaultBool(false),
			mcp.Description("Must be true only after the user confirms the node, instance, and physical PCI device"),
		),
	), handleAttachPCIDevice
}

func handleAttachPCIDevice(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceId, err := requiredParam[string](req, "instance_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// First internal API call: retrieve and validate the instance.
	instance, getErr := api.GetInstanceById(ctx, client, &api.GetInstanceByIdArg{
		InstanceId: instanceId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}
	if instance.Instance.NodeId != nodeId {
		return mcp.NewToolResultError("the selected instance is not hosted on the selected node"), nil
	}
	if instance.Instance.Type != int(enum.InstanceTypeEnumQEMU) {
		return mcp.NewToolResultError("physical PCI devices can only be attached to QEMU instances"), nil
	}

	// Second internal API call: retrieve PCI devices from the instance's node.
	pciInformation, pciErr := api.RetrievePCIDeviceInformationForNode(ctx, client, &api.RetrievePCIDeviceInformationForNodeArg{
		NodeId: nodeId,
	})
	if pciErr != nil {
		return mcp.NewToolResultError(pciErr.Error()), nil
	}

	availableDevices := make([]api.NodePciDevice, 0)
	for _, pciDevice := range *pciInformation {
		if pciDevice.MarkedForPassthrough {
			availableDevices = append(availableDevices, pciDevice)
		}
	}

	if _, selected := req.GetArguments()["pci_slot"]; !selected {
		message := "Select a PCI device from available_pci_devices before attaching it."
		if len(availableDevices) == 0 {
			message = "No PCI devices on the selected node are currently marked for passthrough."
		}
		return mcp.NewToolResultJSON(&attachPCIDeviceOptionsResult{
			Message:             message,
			InstanceId:          instanceId,
			NodeId:              nodeId,
			AvailablePCIDevices: availableDevices,
			NextStep:            "Show these choices to the user. After selection and explicit confirmation, call attach_pci_device again with device_name, the exact pci_slot, and confirm=true. Use is_display and ramfb only when requested.",
		})
	}

	deviceName, err := optionalParam[string](req, "device_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if deviceName == "" {
		return mcp.NewToolResultError("device_name is required when attaching a selected PCI device"), nil
	}
	pciSlot, err := requiredParam[string](req, "pci_slot")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description, err := optionalParam[string](req, "description")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	isDisplay := req.GetBool("is_display", false)
	ramFramebuffer := req.GetBool("ramfb", false)
	if ramFramebuffer && !isDisplay {
		return mcp.NewToolResultError("ramfb can only be enabled when is_display is true"), nil
	}
	if !req.GetBool("confirm", false) {
		return mcp.NewToolResultError("PCI attachment requires explicit user confirmation"), nil
	}

	found := false
	for _, pciDevice := range availableDevices {
		if strings.EqualFold(pciDevice.Slot, pciSlot) {
			found = true
			break
		}
	}
	if !found {
		return mcp.NewToolResultError("the selected PCI device was not found or is not marked for passthrough on the target node"), nil
	}

	response, attachErr := api.AttachDeviceToInstance(ctx, client, &api.AttachDeviceToInstanceArg{
		NodeId:      nodeId,
		InstanceId:  instanceId,
		Name:        deviceName,
		Description: description,
		Type:        api.InstanceDeviceTypePCIDevice,
		Metadata: api.InstanceDeviceMetadata{
			Type: api.InstanceDeviceTypePCIDevice,
			Data: api.PCIDeviceMetadataData{
				Device:         strings.ToLower(pciSlot),
				IsDisplay:      isDisplay,
				RAMFramebuffer: ramFramebuffer,
			},
		},
		Live: req.GetBool("live", false),
	})
	if attachErr != nil {
		return mcp.NewToolResultError(attachErr.Error()), nil
	}

	return mcp.NewToolResultJSON(struct {
		Message     string `json:"message"`
		Name        string `json:"name"`
		NeedsReboot bool   `json:"needs_reboot"`
	}{
		Message:     "PCI device attached successfully",
		Name:        response.Name,
		NeedsReboot: response.NeedsReboot,
	})
}

func ListAllDevicesAttachedToInstance() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_all_devices_attached_to_instance",
		mcp.WithDescription("Retrieve every device currently attached to a specific instance."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Attached Instance Devices",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("Instance whose attached devices should be retrieved"),
		),
	), handleListAllDevicesAttachedToInstance
}

func handleListAllDevicesAttachedToInstance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instanceId, err := requiredParam[string](req, "instance_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	devices, getErr := api.GetAllDevicesAttachedToInstance(
		ctx,
		client,
		&api.GetAllDevicesAttachedToInstanceArg{InstanceId: instanceId},
	)
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}

	return mcp.NewToolResultJSON(devices)
}

func DetachDeviceFromInstance() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("detach_device_from_instance",
		mcp.WithDescription("Detach a named device from an instance. First list attached devices, show the selected device, and obtain explicit confirmation."),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:           "Detach Instance Device",
			ReadOnlyHint:    mcp.ToBoolPtr(false),
			DestructiveHint: mcp.ToBoolPtr(true),
			IdempotentHint:  mcp.ToBoolPtr(false),
		}),
		mcp.WithString("node_id",
			mcp.Required(),
			mcp.Description("Node hosting the instance"),
		),
		mcp.WithString("instance_id",
			mcp.Required(),
			mcp.Description("Instance from which the device will be detached"),
		),
		mcp.WithString("device_name",
			mcp.Required(),
			mcp.Description("Exact attached-device name returned by list_all_devices_attached_to_instance"),
		),
		mcp.WithBoolean("live",
			mcp.DefaultBool(false),
			mcp.Description("Attempt to detach the device while the instance is running"),
		),
		mcp.WithBoolean("confirm",
			mcp.Required(),
			mcp.Description("Must be true after the user confirms the device detachment"),
		),
	), handleDetachDeviceFromInstance
}

func handleDetachDeviceFromInstance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeId, err := requiredParam[string](req, "node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	instanceId, err := requiredParam[string](req, "instance_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	deviceName, err := requiredParam[string](req, "device_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	confirmed, err := req.RequireBool("confirm")
	if err != nil || !confirmed {
		return mcp.NewToolResultError("device detachment requires explicit user confirmation"), nil
	}

	client, err := clientForRequest(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Validate the instance and node inside the same MCP tool call.
	instance, getErr := api.GetInstanceById(ctx, client, &api.GetInstanceByIdArg{
		InstanceId: instanceId,
	})
	if getErr != nil {
		return mcp.NewToolResultError(getErr.Error()), nil
	}
	if instance.Instance.NodeId != nodeId {
		return mcp.NewToolResultError("the selected instance is not hosted on the selected node"), nil
	}

	response, detachErr := api.DetachDeviceFromInstance(
		ctx,
		client,
		&api.DetachDeviceFromInstanceArg{
			NodeId:     nodeId,
			InstanceId: instanceId,
			DeviceName: deviceName,
			Live:       req.GetBool("live", false),
		},
	)
	if detachErr != nil {
		return mcp.NewToolResultError(detachErr.Error()), nil
	}

	return mcp.NewToolResultJSON(struct {
		Message     string `json:"message"`
		NeedsReboot bool   `json:"needs_reboot"`
	}{
		Message:     "Device detached successfully",
		NeedsReboot: response.NeedsReboot,
	})
}
