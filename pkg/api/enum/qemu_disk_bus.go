/*
Copyright 2026 Pextra Inc.

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
package enum

type QEMUDiskBusEnum string

const (
	QEMUDiskBusIDE    QEMUDiskBusEnum = "ide"
	QEMUDiskBusSATA   QEMUDiskBusEnum = "sata"
	QEMUDiskBusSCSI   QEMUDiskBusEnum = "scsi"
	QEMUDiskBusVirtio QEMUDiskBusEnum = "virtio"
	QEMUDiskBusUSB    QEMUDiskBusEnum = "usb"
)

var QEMUDiskBusEnumValues = []QEMUDiskBusEnum{
	QEMUDiskBusIDE,
	QEMUDiskBusSATA,
	QEMUDiskBusSCSI,
	QEMUDiskBusVirtio,
	QEMUDiskBusUSB,
}

func (a QEMUDiskBusEnum) String() string {
	return string(a)
}
