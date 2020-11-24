<h1 align="center">Lumos</h1>

<div align="center">
  <strong>
    Lumos 荧光闪烁
  </strong>
</div>
<br>

<div align="center">
  <a href="https://app.dependabot.com/accounts/jiuzhou-zhao/repos/315576886">
    <img src="https://api.dependabot.com/badges/status?host=github&repo=jiuzhou-zhao/lumos&identifier=315576886" alt="Dependabot">
  </a>
  <img src="https://github.com/jiuzhou-zhao/lumos/workflows/ut/badge.svg?branch=master&event=push" alt="Unit Test">
  <img src="https://github.com/jiuzhou-zhao/lumos/workflows/golangci-lint/badge.svg?branch=master&event=push" alt="GolangCI Linter">
</div>

## About Lumos

`Lumos`这个咒语为“荧光闪烁”，施法时候会在在魔杖尖出现亮光。此后如若前方黑暗，我就为你喊出荧光闪烁吧。 

`Lumos`带人寻找光明。

## 目标 - `HTTP` `HTTPS` 代理

* 支持权限验证
* 支持`relay`模式，中继通过`tls`来保证安全性

## 使用说明

### 单机部署

1. 编译
    ```bash
    git clone git@github.com:jiuzhou-zhao/lumos.git
    cd lumos
    go build -o lumos ./cmd
    ```
2. 配置文件 - `config.yaml`

    ```yml
    Mode: proxy
    ProxyAddress: ":8000"
    Credentials:
      - "u1:p1"
    ```
   > `Credentials`可选，选择了则支持权限认证
   >
  
3. 运行

    ```bash
    ./lumos
    ```
   
### 增加`Local` `Relay` `Proxy` 整套系统

> 参见 `ft` 目录
