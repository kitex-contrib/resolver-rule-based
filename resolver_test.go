/*
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ruleBasedResolver

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/stretchr/testify/assert"
)

// mock struct
var (
	filterFunc1 = func(ctx context.Context, instance []discovery.Instance) []discovery.Instance {
		var res []discovery.Instance
		for _, ins := range instance {
			if ins.Address().Network() == "tcp" {
				res = append(res, ins)
			}
		}
		return res
	}
	filterFunc2 = func(ctx context.Context, instance []discovery.Instance) []discovery.Instance {
		var res []discovery.Instance
		for _, ins := range instance {
			if v, ok := ins.Tag("tag"); ok && v == "1" {
				res = append(res, ins)
			}
		}
		return res
	}
	mockInstances = []discovery.Instance{
		discovery.NewInstance("tcp", "1", 10, map[string]string{"tag": "1"}),
		discovery.NewInstance("tcp", "1", 10, nil),
		discovery.NewInstance("unix", "1", 10, map[string]string{"tag": "1"}),
	}
	mockResolverName = "mockResolver"
	mockResolver     = &discovery.SynthesizedResolver{
		TargetFunc: func(ctx context.Context, target rpcinfo.EndpointInfo) string {
			return target.ServiceName()
		},
		ResolveFunc: func(ctx context.Context, key string) (discovery.Result, error) {
			return discovery.Result{Cacheable: true, CacheKey: key, Instances: mockInstances}, nil
		},
		DiffFunc: func(key string, prev, next discovery.Result) (discovery.Change, bool) {
			// not implemented
			return discovery.Change{}, false
		},
		NameFunc: func() string {
			return mockResolverName
		},
	}
)

func TestFilter(t *testing.T) {
	var res []discovery.Instance
	filter := &instanceFilter{}
	res = filter.filter(context.Background(), mockInstances)
	assert.True(t, reflect.DeepEqual(res, mockInstances))

	rule1 := &FilterRule{
		Name:  "mock_filter_rule1",
		Funcs: []FilterFunc{filterFunc1},
	}

	filter.Rule = rule1
	res = filter.filter(context.Background(), mockInstances)
	assert.Equal(t, 2, len(res))

	rule2 := &FilterRule{
		Name:  "mock_filter_rule2",
		Funcs: []FilterFunc{filterFunc2},
	}
	filter.Rule = rule2
	res = filter.filter(context.Background(), mockInstances)
	assert.Equal(t, 2, len(res))

	rule3 := &FilterRule{
		Name:  "mock_filter_rule3",
		Funcs: []FilterFunc{filterFunc1, filterFunc2},
	}
	filter.Rule = rule3
	res = filter.filter(context.Background(), mockInstances)
	assert.Equal(t, 1, len(res))
}

func TestResolver(t *testing.T) {
	mockRuleName := "filterRule"
	mockServiceName := "mockService"
	ei := rpcinfo.NewEndpointInfo(mockServiceName, "", nil, nil)

	rule := &FilterRule{Name: mockRuleName, Funcs: []FilterFunc{filterFunc1}}
	rbr := NewRuleBasedResolver(mockResolver, rule)
	// target
	target := rbr.Target(context.Background(), ei)
	assert.Equal(t, mockServiceName, target)

	// resolve
	res, err := rbr.Resolve(context.Background(), target)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(res.Instances))
	// name
	assert.Equal(t, fmt.Sprintf("%s|%s", mockResolverName, mockRuleName), rbr.Name())
}
