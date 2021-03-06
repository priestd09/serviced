// Copyright 2016 The Serviced Authors.
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
package node

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"time"

	"github.com/control-center/serviced/domain/service"
	"github.com/control-center/serviced/rpc/master"
	"github.com/control-center/serviced/utils/cache"
)

type ServiceCache struct {
	master string // the connection string to the master agent
	cache  cache.LRUCache

	masterClient master.ClientInterface // ONLY USED FOR UNIT-TESTING
}

type cachedService struct {
	Service          *service.Service
	TenantID         string
	ServiceNamePath  string
}

func NewServiceCache(master string) *ServiceCache {
	serviceCache := ServiceCache{
		master: master,
	}

	maxCacheSize := 60
	itemTimeToLive := time.Minute * 2
	cleanupInterval := time.Second * 30
	serviceCache.cache, _ = cache.NewSimpleLRUCache(maxCacheSize, itemTimeToLive, cleanupInterval, nil)
	return &serviceCache
}

func (sc *ServiceCache) GetEvaluatedService(serviceID string, instanceID int) (*service.Service, string, string, error) {
	logger := plog.WithFields(log.Fields{
		"serviceid":  serviceID,
		"instanceid": instanceID,
	})

	var item cachedService
	key := fmt.Sprintf("%s-%d", serviceID, instanceID)
	data, ok := sc.cache.Get(key)
	if ok {
		item, _ = data.(cachedService)
		return item.Service, item.TenantID, item.ServiceNamePath, nil
	}

	masterClient, err := sc.getMasterClient()
	if err != nil {
		logger.WithField("master", sc.master).WithError(err).Error("Could not connect to the master")
		return nil, "", "", err
	}
	defer masterClient.Close()

	item.Service, item.TenantID, item.ServiceNamePath, err = masterClient.GetEvaluatedService(serviceID, instanceID)
	if err != nil {
		logger.WithError(err).Error("Failed to get service")
		return nil, "", "", err
	}
	sc.cache.Set(key, item)
	return item.Service, item.TenantID, item.ServiceNamePath, nil
}

func (sc *ServiceCache) Invalidate(serviceID string, instanceID int) {
	key := fmt.Sprintf("%s-%d", serviceID, instanceID)
	sc.cache.Invalidate(key)
}

func (sc *ServiceCache) getMasterClient() (master.ClientInterface, error) {
	if sc.masterClient != nil {
		return sc.masterClient, nil // ONLY USED FOR UNIT-TESTING
	}

	return master.NewClient(sc.master)
}
