// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package schedule

import (
	"github.com/pingcap/pd/server/core"
)

// BasicCluster provides basic data member and interface for a tikv cluster.
type BasicCluster struct {
	Stores  *core.StoresInfo
	Regions *core.RegionsInfo
}

// NewBasicCluster creates a BasicCluster.
func NewBasicCluster() *BasicCluster {
	return &BasicCluster{
		Stores:  core.NewStoresInfo(),
		Regions: core.NewRegionsInfo(),
	}
}

// GetStores returns all Stores in the cluster.
func (bc *BasicCluster) GetStores() []*core.StoreInfo {
	return bc.Stores.GetStores()
}

// GetStore searches for a store by ID.
func (bc *BasicCluster) GetStore(storeID uint64) *core.StoreInfo {
	return bc.Stores.GetStore(storeID)
}

// GetRegion searches for a region by ID.
func (bc *BasicCluster) GetRegion(regionID uint64) *core.RegionInfo {
	return bc.Regions.GetRegion(regionID)
}

// GetRegionStores returns all Stores that contains the region's peer.
func (bc *BasicCluster) GetRegionStores(region *core.RegionInfo) []*core.StoreInfo {
	var Stores []*core.StoreInfo
	for id := range region.GetStoreIds() {
		if store := bc.Stores.GetStore(id); store != nil {
			Stores = append(Stores, store)
		}
	}
	return Stores
}

// GetFollowerStores returns all Stores that contains the region's follower peer.
func (bc *BasicCluster) GetFollowerStores(region *core.RegionInfo) []*core.StoreInfo {
	var Stores []*core.StoreInfo
	for id := range region.GetFollowers() {
		if store := bc.Stores.GetStore(id); store != nil {
			Stores = append(Stores, store)
		}
	}
	return Stores
}

// GetLeaderStore returns all Stores that contains the region's leader peer.
func (bc *BasicCluster) GetLeaderStore(region *core.RegionInfo) *core.StoreInfo {
	return bc.Stores.GetStore(region.GetLeader().GetStoreId())
}

// GetAdjacentRegions returns region's info that is adjacent with specific region
func (bc *BasicCluster) GetAdjacentRegions(region *core.RegionInfo) (*core.RegionInfo, *core.RegionInfo) {
	return bc.Regions.GetAdjacentRegions(region)
}

// BlockStore stops balancer from selecting the store.
func (bc *BasicCluster) BlockStore(storeID uint64) error {
	return bc.Stores.BlockStore(storeID)
}

// UnblockStore allows balancer to select the store.
func (bc *BasicCluster) UnblockStore(storeID uint64) {
	bc.Stores.UnblockStore(storeID)
}

// SetStoreOverload stops balancer from selecting the store.
func (bc *BasicCluster) SetStoreOverload(storeID uint64) {
	bc.Stores.SetStoreOverload(storeID)
}

// ResetStoreOverload allows balancer to select the store.
func (bc *BasicCluster) ResetStoreOverload(storeID uint64) {
	bc.Stores.ResetStoreOverload(storeID)
}

// RandFollowerRegion returns a random region that has a follower on the store.
func (bc *BasicCluster) RandFollowerRegion(storeID uint64, opts ...core.RegionOption) *core.RegionInfo {
	return bc.Regions.RandFollowerRegion(storeID, opts...)
}

// RandLeaderRegion returns a random region that has leader on the store.
func (bc *BasicCluster) RandLeaderRegion(storeID uint64, opts ...core.RegionOption) *core.RegionInfo {
	return bc.Regions.RandLeaderRegion(storeID, opts...)
}

// GetAverageRegionSize returns the average region approximate size.
func (bc *BasicCluster) GetAverageRegionSize() int64 {
	return bc.Regions.GetAverageRegionSize()
}

// PutStore put a store
func (bc *BasicCluster) PutStore(store *core.StoreInfo) {
	bc.Stores.SetStore(store)
}

// DeleteStore deletes a store
func (bc *BasicCluster) DeleteStore(store *core.StoreInfo) {
	bc.Stores.DeleteStore(store)
}

// PutRegion put a region
func (bc *BasicCluster) PutRegion(region *core.RegionInfo) {
	bc.Regions.SetRegion(region)
}
