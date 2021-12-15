<h1 align="center">Lumos</h1>

<div align="center">
  <strong>
    Lumos 荧光闪烁
  </strong>
</div>
<br>

<div align="center">
  <img src="https://github.com/jiuzhou-zhao/lumos/workflows/ut/badge.svg?branch=master&event=push" alt="Unit Test">
  <img src="https://github.com/jiuzhou-zhao/lumos/workflows/golangci-lint/badge.svg?branch=master&event=push" alt="GolangCI Linter">
</div>

## About Lumos

`Lumos`这个咒语为“荧光闪烁”，施法时候会在魔杖尖出现亮光。此后如若前方黑暗，我就为你喊出荧光闪烁吧！

`Lumos`带人寻找光明。

## 目标 - `HTTP` `Socks5[TCP]` 代理

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
    DialTimeout: 30s
    Credentials:
      - "u1:p1"
    ```
   > `Credentials`可选，选择了则支持权限认证
   > 
   > `Mode` 设置为 `socks5` 则为 `socks5`代理

3. 运行

    ```bash
    ./lumos
    ```

### 增加`Local` `Relay` `Proxy` 整套系统

1. 克隆代码，编译 或者 直接下载`Release`中的二进制包 - 获取 `lumos` 可执行程序 和 `scripts`生成证书的工具链
    ```bash
    git clone git@github.com:jiuzhou-zhao/lumos.git
    cd lumos
    go build -o lumos ./cmd
    ```

2. 生成证书
    ```bash
    ./scripts/certs.sh
    ```
   > 如果已经有证书，则不用此步骤
   >
   > 生成的证书会存放在`certs`目录里
   >
   > 注意：修改脚本中目录`server.conf`中的`alt_names` 字段来适配真正的域名
   >
   > 所以，如果有多个服务器，要么为每个服务器都部署不同的证书，要么把所有域名 IP都加入`alt_names`字段里
   
3. 拷贝`config-sample.yaml`为`config.yaml`, 修改 - 参见 `ft` 目录

4. 分别在`local`, `relay`, `server` 上部署配置文件和`lumos`程序, 其中`relay`可以有多个

    * 配置文件中`Proxy`可以取值 `local`, `relay`, `socks5`, `proxy`
    * 数据流为 `浏览器` <-> `local` <-> `relay` <-> `relay` <-> `proxy server`[http or proxy]

4. 各个节点执行命令
    ```bash
    ./lumos
    ```

#### 配置文件模板 [`local`+`proxy`]

##### `local` - 例如`sv1`上配置的，`relay`在`sv2`上
```yaml
Mode: local
ProxyAddress: ":8000"
RemoteProxyAddress: "sv2:8001"
Secure:
  TLSEnableFlag:
    ConnectServerUseTLS: true
  ConnectServerTLSConfig:
    Cert: ./certs/proxy-client.crt
    Key: ./certs/proxy-client.key
    RootCAs:
      - ./certs/ca.crt
      - ./certs/server.crt

DialTimeout: 30s
```

##### `relay` - 例如`sv2`上配置的，`proxy`在`sv3`上
```yaml
Mode: relay
ProxyAddress: ":8001"
RemoteProxyAddress: "sv3:8001"
Secure:
  TLSEnableFlag:
    ConnectServerUseTLS: true
    ServerUseTLS: true
  ConnectServerTLSConfig:
    Cert: ./certs/proxy-client.crt
    Key: ./certs/proxy-client.key
    RootCAs:
      - ./certs/ca.crt
      - ./certs/server.crt
  ServerTLSConfig:
    Cert: ./certs/proxy-server.crt
    Key: ./certs/proxy-server.key
    RootCAs:
      - ./certs/ca.crt
      - ./certs/client.crt

DialTimeout: 30s
```

##### `http proxy` - 例如`sv3`上配置的
```yaml
Mode: socks5
ProxyAddress: ":8001"
Secure:
  TLSEnableFlag:
    ServerUseTLS: true
  ServerTLSConfig:
    Cert: ./certs/proxy-server.crt
    Key: ./certs/proxy-server.key
    RootCAs:
      - ./certs/ca.crt
      - ./certs/client.crt

DialTimeout: 30s
```

> `Mode` 可改为 `socks5` 来变为 `socks5`代理
>
> 别忘记防火墙
>