package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	etcd "github.com/kitex-contrib/registry-etcd"

	ruleBasedResolver "github.com/kitex-contrib/resolver-rule-based"
)

var (
	etcdAddr     = "127.0.0.1:2379"
	serviceName  = "service"
	tagKey       = "k"
	tagValues    = []string{"v1", "v2"}
	instanceTags = []map[string]string{
		{tagKey: tagValues[0]},
		{tagKey: tagValues[1]},
	}
)

func resolve() {
	// use etcd resolver
	etcdResolver, err := etcd.NewEtcdResolver([]string{etcdAddr})
	if err != nil {
		panic(err)
	}
	filterFunc := func(ctx context.Context, instance []discovery.Instance) []discovery.Instance {
		var res []discovery.Instance
		for _, ins := range instance {
			if v, ok := ins.Tag(tagKey); ok && v == tagValues[0] {
				// only match tag with {tagKey: tagValues[0]}
				res = append(res, ins)
			}
		}
		return res
	}
	// Construct the filterRule
	filterRule := &ruleBasedResolver.FilterRule{Name: "rule-name", Funcs: []ruleBasedResolver.FilterFunc{filterFunc}}
	// build rule based resolver
	rbr := ruleBasedResolver.NewRuleBasedResolver(etcdResolver, filterRule)

	// service discovery
	ctx := context.Background()
	ei := rpcinfo.NewEndpointInfo(serviceName, "", nil, nil)
	desc := rbr.Target(ctx, ei)
	res, err := rbr.Resolve(ctx, desc)
	if err != nil {
		panic(err)
	}
	// the instance should match the filter rule
	v, _ := res.Instances[0].Tag(tagKey)
	fmt.Println(fmt.Sprintf("[Resolver]: get instance with tag, [%s:%s]", tagKey, v))
}

func main() {
	// instances
	var instances []*registry.Info
	for i := 0; i < len(instanceTags); i++ {
		addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", strconv.Itoa(8888+i)))
		instances = append(instances, &registry.Info{
			ServiceName: serviceName,
			Addr:        addr,
			Tags:        instanceTags[i],
		})
	}

	// register
	r, err := etcd.NewEtcdRegistry([]string{etcdAddr})
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(instances); i++ {
		err := r.Register(instances[i])
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		for i := 0; i < len(instances); i++ {
			r.Deregister(instances[i])
		}
	}()

	// Test the resolver
	resolve()
}
