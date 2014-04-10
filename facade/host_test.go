// Copyright 2014, The Serviced Authors. All rights reserved.
// Use of this source code is governed by a
// license that can be found in the LICENSE file.

package facade

import (
	"github.com/zenoss/glog"
	"github.com/zenoss/serviced/datastore"
	"github.com/zenoss/serviced/datastore/elastic"
	"github.com/zenoss/serviced/domain/host"
	. "gopkg.in/check.v1"

	"github.com/zenoss/serviced/datastore/context"
	"github.com/zenoss/serviced/domain/pool"
	"testing"
)

// This plumbs gocheck into testing
func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&S{
	ElasticTest: elastic.ElasticTest{
		Index: "controlplane",
		Mappings: map[string]string{
			"host":         "../domain/host/host_mapping.json",
			"resourcepool": "../domain/pool/pool_mapping.json",
		},
	}})

type S struct {
	elastic.ElasticTest
	ctx context.Context
	tf  *Facade
}

func (s *S) SetUpTest(c *C) {
	context.Register(s.Driver())
	s.ctx = context.Get()
	s.tf = New()
}

func (s *S) Test_HostCRUD(t *C) {
	testid := "facadetestid"
	defer s.tf.RemoveHost(s.ctx, testid)

	//create pool for test
	pool := pool.New("pool-id")
	if err := s.tf.AddResourcePool(s.ctx, pool); err != nil {
		t.Fatalf("Could not add pool for test: %v", err)
	}

	//fill host with required values
	h, err := host.Build("", pool.ID, []string{}...)
	h.ID = "facadetestid"
	if err != nil {
		t.Fatalf("Unexpected error building host: %v", err)
	}
	glog.Infof("Facade test add host %v", h)
	err = s.tf.AddHost(s.ctx, h)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	//Test re-add fails
	err = s.tf.AddHost(s.ctx, h)
	if err == nil {
		t.Errorf("Expected already exists error: %v", err)
	}

	h2, err := s.tf.GetHost(s.ctx, testid)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if h2 == nil {
		t.Error("Unexpected nil host")

	} else if !host.HostEquals(t, h, h2) {
		t.Error("Hosts did not match")
	}

	//Test update
	h.Memory = 1024
	err = s.tf.UpdateHost(s.ctx, h)
	h2, err = s.tf.GetHost(s.ctx, testid)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !host.HostEquals(t, h, h2) {
		t.Error("Hosts did not match")
	}

	//test delete
	err = s.tf.RemoveHost(s.ctx, testid)
	h2, err = s.tf.GetHost(s.ctx, testid)
	if err != nil && !datastore.IsErrNoSuchEntity(err) {
		t.Errorf("Unexpected error: %v", err)
	}

}

/*
func Test_GetHosts(t *testing.T) {
	if tf == nil {
		t.Fatalf("Test failed to initialize")
	}
	hid1 := "gethosts1"
	hid2 := "gethosts2"

	defer s.tf.RemoveHost(s.ctx, hid1)
	defer s.tf.RemoveHost(s.ctx, hid2)

	host, err := host.Build("", "pool-id", []string{}...)
	host.Id = "Test_GetHosts1"
	if err != nil {
		t.Fatalf("Unexpected error building host: %v", err)
	}
	err = hs.Put(s.ctx, host)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	time.Sleep(1000 * time.Millisecond)
	hosts, err := hs.GetUpTo(s.ctx, 1000)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if len(hosts) != 1 {
		t.Errorf("Expected %v results, got %v :%v", 1, len(hosts), hosts)
	}

	host.Id = "Test_GetHosts2"
	err = hs.Put(s.ctx, host)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	time.Sleep(1000 * time.Millisecond)
	hosts, err = hs.GetUpTo(s.ctx, 1000)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if len(hosts) != 2 {
		t.Errorf("Expected %v results, got %v :%v", 2, len(hosts), hosts)
	}

}
*/
