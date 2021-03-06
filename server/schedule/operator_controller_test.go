// Copyright 2018 PingCAP, Inc.
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
	"container/heap"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/kvproto/pkg/pdpb"
)

var _ = Suite(&testOperatorControllerSuite{})

type testOperatorControllerSuite struct{}

// issue #1338
func (t *testOperatorControllerSuite) TestGetOpInfluence(c *C) {
	opt := NewMockSchedulerOptions()
	tc := NewMockCluster(opt)
	oc := NewOperatorController(tc, nil)
	tc.AddLeaderStore(2, 1)
	tc.AddLeaderRegion(1, 1, 2)
	tc.AddLeaderRegion(2, 1, 2)
	steps := []OperatorStep{
		RemovePeer{FromStore: 2},
	}
	op1 := NewOperator("test", 1, &metapb.RegionEpoch{}, OpRegion, steps...)
	op2 := NewOperator("test", 2, &metapb.RegionEpoch{}, OpRegion, steps...)
	oc.SetOperator(op1)
	oc.SetOperator(op2)
	go func() {
		for {
			oc.RemoveOperator(op1)
		}
	}()
	go func() {
		for {
			oc.GetOpInfluence(tc)
		}
	}()
	time.Sleep(1 * time.Second)
	c.Assert(oc.GetOperator(2), NotNil)
}

func (t *testOperatorControllerSuite) TestOperatorStatus(c *C) {
	opt := NewMockSchedulerOptions()
	tc := NewMockCluster(opt)
	oc := NewOperatorController(tc, MockHeadbeatStream{})
	tc.AddLeaderStore(1, 2)
	tc.AddLeaderStore(2, 0)
	tc.AddLeaderRegion(1, 1, 2)
	tc.AddLeaderRegion(2, 1, 2)
	steps := []OperatorStep{
		RemovePeer{FromStore: 2},
		AddPeer{ToStore: 2, PeerID: 4},
	}
	op1 := NewOperator("test", 1, &metapb.RegionEpoch{}, OpRegion, steps...)
	op2 := NewOperator("test", 2, &metapb.RegionEpoch{}, OpRegion, steps...)
	region1 := tc.GetRegion(1)
	region2 := tc.GetRegion(2)
	oc.SetOperator(op1)
	oc.SetOperator(op2)
	c.Assert(oc.GetOperatorStatus(1).Status, Equals, pdpb.OperatorStatus_RUNNING)
	c.Assert(oc.GetOperatorStatus(2).Status, Equals, pdpb.OperatorStatus_RUNNING)
	op1.createTime = time.Now().Add(-10 * time.Minute)
	region2 = tc.ApplyOperatorStep(region2, op2)
	tc.PutRegion(region2)
	oc.Dispatch(region1, "test")
	oc.Dispatch(region2, "test")
	c.Assert(oc.GetOperatorStatus(1).Status, Equals, pdpb.OperatorStatus_TIMEOUT)
	c.Assert(oc.GetOperatorStatus(2).Status, Equals, pdpb.OperatorStatus_RUNNING)
	tc.ApplyOperator(op2)
	oc.Dispatch(region2, "test")
	c.Assert(oc.GetOperatorStatus(2).Status, Equals, pdpb.OperatorStatus_SUCCESS)
}

func (t *testOperatorControllerSuite) TestPollDispatchRegion(c *C) {
	opt := NewMockSchedulerOptions()
	tc := NewMockCluster(opt)
	oc := NewOperatorController(tc, MockHeadbeatStream{})
	tc.AddLeaderStore(1, 2)
	tc.AddLeaderStore(2, 0)
	tc.AddLeaderRegion(1, 1, 2)
	tc.AddLeaderRegion(2, 1, 2)
	steps := []OperatorStep{
		RemovePeer{FromStore: 2},
		AddPeer{ToStore: 2, PeerID: 4},
	}
	op1 := NewOperator("test", 1, &metapb.RegionEpoch{}, OpRegion, TransferLeader{ToStore: 2})
	op2 := NewOperator("test", 2, &metapb.RegionEpoch{}, OpRegion, steps...)
	region1 := tc.GetRegion(1)
	region2 := tc.GetRegion(2)
	// Adds operator and pushes to the notifier queue.
	{
		oc.SetOperator(op1)
		oc.SetOperator(op2)
		heap.Push(&oc.opNotifierQueue, &operatorWithTime{op: op1, time: time.Now().Add(100 * time.Millisecond)})
		heap.Push(&oc.opNotifierQueue, &operatorWithTime{op: op2, time: time.Now().Add(500 * time.Millisecond)})
	}
	// fisrt poll got nil
	r, next := oc.pollNeedDispatchRegion()
	c.Assert(r, IsNil)
	c.Assert(next, IsFalse)

	// after wait 100 millisecond, the region1 need to dispatch, but not region2.
	time.Sleep(100 * time.Millisecond)
	r, next = oc.pollNeedDispatchRegion()
	c.Assert(r, NotNil)
	c.Assert(next, IsTrue)
	c.Assert(r.GetID(), Equals, region1.GetID())
	r, next = oc.pollNeedDispatchRegion()
	c.Assert(r, IsNil)
	c.Assert(next, IsFalse)

	// after waiting 500 millseconds, the region2 need to dispatch
	time.Sleep(400 * time.Millisecond)
	r, next = oc.pollNeedDispatchRegion()
	c.Assert(r, NotNil)
	c.Assert(next, IsTrue)
	c.Assert(r.GetID(), Equals, region2.GetID())
}
