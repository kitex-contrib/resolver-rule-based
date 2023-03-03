# Kitex Rule Based Resolver
This project provides a rule-based resolver for Kitex. It allows user to configure rules to filter service instances in Service Discovery, achieving the function of traffic split.

This resolver needs an implemented Resolver, which is able to resolve instances from the registry center, and some customized filter rules (e.g. filter by tags in the instances).

## Usage
1. Implement your own Resolver. Refer to this doc about the definition of Resolver: [Service Discovery Extension](https://www.cloudwego.io/docs/kitex/tutorials/framework-exten/service_discovery/)

2. Define filter rules. 

```
// Define a filter function.
// For example, only get the instances with a tag of {"k":"v"}.
filterFunc = func(ctx context.Context, instance []discovery.Instance) []discovery.Instance {
     var res []discovery.Instance
     for _, ins := range instance {
         if v, ok := ins.Tag("k"); ok && v == "v" {
             res = append(res, ins)
         }
     }
     return res
}
// Construct the filterRule
filterRule := &FilterRule{Name: "rule-name", Funcs: []FilterFunc{filterFunc}} 
```
Notice: the FilterFuncs will be executed sequentially.

3. Configure the resolver
```
import (
   ruleBasedResolver "github.com/kitex-contrib/resolver-rule-based"
   "github.com/cloudwego/kitex/client"
   "github.com/cloudwego/kitex/pkg/discovery"
)

// implement your resolver
var newResolver discovery.Resolver

// construct a RuleBasedResolver with the `newResolver` and `filterRule`
tagResolver := ruleBasedResolver.NewRuleBasedResolver(resolver, filterRule)

// add this option when construct Kitex Client
opt := client.WithResolver(tagResolver) 
```
