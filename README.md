# data 数据库ORM组件
> [文档：https://farseer-go.github.io/doc/](https://farseer-go.github.io/doc/)

> 包：`"github.com/farseer-go/data"`

> 模块：`data.Module`

## 概述
data组件提供数据库ORM操作，将数据库多张表组织到一个`上下文`中。并使用统一的`./farseer.yaml`配置

?> 目前orm底层的组件使用的是gorm，data组件主要为了做进一步的封装，使得我们在使用时更加简单易用。

data组件，采用数据库上下文的概念，将多个model组合在一起，方便统一管理。