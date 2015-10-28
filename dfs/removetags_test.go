// Copyright 2015 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build unit

package dfs_test

import (
	"github.com/control-center/serviced/volume"
	volumemocks "github.com/control-center/serviced/volume/mocks"
	. "gopkg.in/check.v1"
)

func (s *DFSTestSuite) TestRemoveTags_FailGetVolume(c *C) {
	snapshotID := "SnapID"
	tags := []string{"tag1"}
	s.disk.On("GetTenant", snapshotID).Return(&volumemocks.Volume{}, volume.ErrVolumeNotExists).Once()
	newTagList, err := s.dfs.RemoveTags(snapshotID, tags)
	c.Assert(newTagList, IsNil)
	c.Assert(err, Equals, volume.ErrVolumeNotExists)
}

func (s *DFSTestSuite) TestRemoveTags_FailRemoveTags(c *C) {
	snapshotID := "SnapID"
	tags := []string{"tag1"}

	vol := &volumemocks.Volume{}

	s.disk.On("GetTenant", snapshotID).Return(vol, nil)
	vol.On("RemoveSnapshotTags", snapshotID, tags).Return(nil, ErrTestRemoveSnapshotTagsFailed).Once()

	newTagList, err := s.dfs.RemoveTags(snapshotID, tags)
	c.Assert(newTagList, IsNil)
	c.Assert(err, Equals, ErrTestRemoveSnapshotTagsFailed)
}

func (s *DFSTestSuite) TestRemoveTags_Success(c *C) {
	snapshotID := "SnapID"
	tags := []string{"tag1", "tag2"}
	expectedTags := []string{"tag3"}

	vol := &volumemocks.Volume{}

	s.disk.On("GetTenant", snapshotID).Return(vol, nil)
	vol.On("RemoveSnapshotTags", snapshotID, tags).Return(expectedTags, nil).Once()

	newTagList, err := s.dfs.RemoveTags(snapshotID, tags)
	c.Assert(newTagList, DeepEquals, expectedTags)
	c.Assert(err, IsNil)
}