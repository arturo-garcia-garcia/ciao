// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/01org/ciao/ciao-controller/types"
	"github.com/01org/ciao/payloads"
	"github.com/01org/ciao/ssntp/uuid"
)

func validateVMWorkload(req types.Workload) error {
	// FWType must be either EFI or legacy.
	if req.FWType != string(payloads.EFI) && req.FWType != payloads.Legacy {
		return types.ErrBadRequest
	}

	// Must have storage for VMs
	if len(req.Storage) == 0 {
		return types.ErrBadRequest
	}

	return nil
}

func validateContainerWorkload(req types.Workload) error {
	// we should reject anything with ImageID set, but
	// we'll just ignore it.
	if req.ImageName == "" {
		return types.ErrBadRequest
	}

	return nil
}

func validateWorkloadStorage(req types.Workload) error {
	bootableCount := 0
	for i := range req.Storage {
		// check that a workload type is specified
		if req.Storage[i].SourceType == "" {
			return types.ErrBadRequest
		}

		// you may not request a sized volume unless it's empty.
		if req.Storage[i].Size > 0 && req.Storage[i].SourceType != types.Empty {
			return types.ErrBadRequest
		}

		// you may not request a bootable empty volume.
		if req.Storage[i].Bootable && req.Storage[i].SourceType == types.Empty {
			return types.ErrBadRequest
		}

		if req.Storage[i].ID != "" {
			// validate that the id is at least valid
			// uuid4.
			_, err := uuid.Parse(req.Storage[i].ID)
			if err != nil {
				return types.ErrBadRequest
			}

			// If we have an ID we must have a type to get it from
			if req.Storage[i].SourceType != types.Empty {
				return types.ErrBadRequest
			}
		} else {
			if req.Storage[i].SourceType == types.Empty {
				// you many not use a source ID with empty.
				if req.Storage[i].SourceID != "" {
					return types.ErrBadRequest
				}
			} else {
				// you must specify an ID with volume/image
				if req.Storage[i].SourceID == "" {
					return types.ErrBadRequest
				}
			}
		}

		if req.Storage[i].Bootable {
			bootableCount++
		}
	}

	// must be at least one bootable volume
	if req.VMType == payloads.QEMU && bootableCount == 0 {
		return types.ErrBadRequest
	}

	return nil
}

// this is probably an insufficient amount of checking.
func validateWorkloadRequest(req types.Workload) error {
	// ID must be blank.
	if req.ID != "" {
		return types.ErrBadRequest
	}

	if req.VMType == payloads.QEMU {
		err := validateVMWorkload(req)
		if err != nil {
			return err
		}
	} else {
		err := validateContainerWorkload(req)
		if err != nil {
			return err
		}
	}

	if req.ImageID != "" {
		// validate that the image id is at least valid
		// uuid4.
		_, err := uuid.Parse(req.ImageID)
		if err != nil {
			return types.ErrBadRequest
		}
	}

	if req.Config == "" {
		return types.ErrBadRequest
	}

	if len(req.Storage) > 0 {
		err := validateWorkloadStorage(req)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) CreateWorkload(req types.Workload) (types.Workload, error) {
	err := validateWorkloadRequest(req)
	if err != nil {
		return req, err
	}

	// create a workload storage resource for this new workload.
	if req.ImageID != "" {
		// validate that the image id is at least valid
		// uuid4.
		_, err = uuid.Parse(req.ImageID)
		if err != nil {
			return req, err
		}

		storage := types.StorageResource{
			Bootable:   true,
			Ephemeral:  true,
			SourceType: types.ImageService,
			SourceID:   req.ImageID,
		}

		req.ImageID = ""
		req.Storage = append(req.Storage, storage)
	}

	req.ID = uuid.Generate().String()

	err = c.ds.AddWorkload(req)
	return req, err
}
