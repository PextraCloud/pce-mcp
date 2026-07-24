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

type VolumeStatusEnum int

const (
	VolumeStatusEnumCreating   VolumeStatusEnum = iota // 0
	VolumeStatusEnumAvailable                          // 1
	VolumeStatusEnumAttaching                          // 2
	VolumeStatusEnumDetaching                          // 3
	VolumeStatusEnumDestroying                         // 4
	VolumeStatusEnumResizing                           // 5
	VolumeStatusEnumError                              // 6
)

func (e VolumeStatusEnum) String() string {
	return [...]string{"Creating", "Available", "Attaching", "Detaching", "Destroying", "Resizing", "Error"}[e]
}
