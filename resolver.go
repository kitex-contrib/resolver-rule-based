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

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// FilterRule defines the rules that will be used to filter service mockInstances.
type FilterRule struct {
	Funcs []FilterFunc
	Name  string // the name of the filter rule should be unique. It will be used for cache key.
}

// FilterFunc input original mockInstances and return mockInstances after filtering.
type FilterFunc func(ctx context.Context, instances []discovery.Instance) []discovery.Instance

type instanceFilter struct {
	Rule *FilterRule
}

func (filter *instanceFilter) filter(ctx context.Context, instances []discovery.Instance) []discovery.Instance {
	if len(instances) == 0 {
		return instances
	}
	if filter.Rule == nil || len(filter.Rule.Funcs) == 0 {
		return instances
	}
	for _, f := range filter.Rule.Funcs {
		instances = f(ctx, instances)
	}
	return instances
}

func (filter *instanceFilter) name() string {
	if filter.Rule == nil {
		return ""
	}
	return filter.Rule.Name
}

// NewRuleBasedResolver constructs a RuleBasedResolver with the input resolver and filterRule.
func NewRuleBasedResolver(resolver discovery.Resolver, rule *FilterRule) *RuleBasedResolver {
	if resolver == nil || rule == nil {
		panic("Resolver and FilterRule should be provided")
	}
	return &RuleBasedResolver{resolver, &instanceFilter{rule}}
}

var _ discovery.Resolver = &RuleBasedResolver{}

// RuleBasedResolver implements the discovery.Resolver interface.
// It wraps the resolver and filter that enable rule-based instance filter in Service Discovery.
type RuleBasedResolver struct {
	resolver discovery.Resolver
	filter   *instanceFilter
}

// Target implements the discovery.Resolver interface.
func (c *RuleBasedResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return c.resolver.Target(ctx, target)
}

// Resolve implements the discovery.Resolver interface.
func (c *RuleBasedResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	res, err := c.resolver.Resolve(ctx, desc)
	if err != nil {
		return discovery.Result{}, err
	}
	res.Instances = c.filter.filter(ctx, res.Instances)
	return res, nil
}

// Diff implements the discovery.Resolver interface.
func (c *RuleBasedResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return c.resolver.Diff(cacheKey, prev, next)
}

// Name implements the discovery.Resolver interface
func (c *RuleBasedResolver) Name() string {
	return fmt.Sprintf("%s|%s", c.resolver.Name(), c.filter.name())
}
