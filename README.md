# data 数据库ORM组件

> 包：`"github.com/farseer-go/data"`
> 
> 模块：`data.Module`

- `Document`
    - [English](https://farseer-go.gitee.io/en-us/)
    - [中文](https://farseer-go.gitee.io/)
    - [English](https://farseer-go.github.io/doc/en-us/)
- Source
    - [github](https://github.com/farseer-go/fs)

![](https://img.shields.io/github/stars/farseer-go?style=social)
![](https://img.shields.io/github/license/farseer-go/data)
![](https://img.shields.io/github/go-mod/go-version/farseer-go/data)
![](https://img.shields.io/github/v/release/farseer-go/data)
![go-version](https://img.shields.io/github/go-mod/go-version/farseer-go/data)
![](https://img.shields.io/github/languages/code-size/farseer-go/data)
[![Build](https://github.com/farseer-go/data/actions/workflows/go.yml/badge.svg)](https://github.com/farseer-go/data/actions/workflows/go.yml)
![](https://goreportcard.com/badge/github.com/farseer-go/data)

## 概述

data组件提供数据库ORM操作，将数据库多张表组织到一个`上下文`中。并使用统一的`./farseer.yaml`配置

> 目前orm底层的组件使用的是gorm，data组件主要为了做进一步的封装，使得我们在使用时更加简单易用。

data组件，采用数据库上下文的概念，将多个model组合在一起，方便统一管理。