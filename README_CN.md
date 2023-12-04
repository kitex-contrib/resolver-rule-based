# 基于规则的 Kitex 解析器

[English](README.md) | 中文

This project provides a rule-based resolver for Kitex. It allows user to configure rules to filter service instances in Service Discovery, achieving the function of traffic split.

This resolver needs an implemented Resolver, which is able to resolve instances from the registry center, and some customized filter rules (e.g. filter by tags in the instances).

这个项目为 Kitex 提供了一个基于规则的解析器。它允许用户在服务发现中配置规则来过滤服务实例，实现流量切分的功能。

这个解析器需要一个已实现的 Resolver，能够从注册中心解析实例，同时还需要一些定制化的过滤规则（例如，在实例中根据标签进行过滤）。

## 用法
1. 实现你自己的解析器。参考这个文档关于解析器的定义： [服务发现扩展](https://www.cloudwego.io/zh/docs/kitex/tutorials/framework-exten/service_discovery/)

2. 定义过滤规则。

    ```go
    // 定义一个过滤函数
    // 例如，只获取具有标签 {"k":"v"} 的实例
    filterFunc := func(ctx context.Context, instance []discovery.Instance) []discovery.Instance {
         var res []discovery.Instance
         for _, ins := range instance {
             if v, ok := ins.Tag("k"); ok && v == "v" {
                 res = append(res, ins)
             }
         }
         return res
    }
    // 构造过滤规则
    filterRule := &FilterRule{Name: "rule-name", Funcs: []FilterFunc{filterFunc}} 
    ```
    注意：FilterFuncs 将按顺序执行。

3. 配置解析器

    ```go
    import (
       ruleBasedResolver "github.com/kitex-contrib/resolver-rule-based"
       "github.com/cloudwego/kitex/client"
       "github.com/cloudwego/kitex/pkg/discovery"
    )
    
    // 实现你的解析器
    var newResolver discovery.Resolver
    
    // 使用 `newResolver` 和 `filterRule` 构造一个 RuleBasedResolver
    tagResolver := ruleBasedResolver.NewRuleBasedResolver(resolver, filterRule)
    
    // 在构建 Kitex 客户端时添加此选项
    opt := client.WithResolver(tagResolver) 
    ```

## 示例
请参考 `/demo` 获取详情。