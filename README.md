# v2scar
side car for V2ray

## 原理介绍

该项目需要以`side car`的形式和`v2ray`部署在是一台机器上

他通过开放的grpc接口动态添加/删除`Vmess`,并且能统计用户的流量

通过接入`django-sspanel`的api，可以动态调节v2ray的Vmess用户

目前实现了以下几个接口：

* AddInboundUser
* RemoveInboundUser
* GetAndResetUserTraffic

## 使用说明

### 配置环境变量

```bash
export V2SCAR_API_ENDPOINT="xxxx" # 这个是django-sspanel的sync api 地址
export V2SCAR_GRPC_ENDPOINT="127.0.0.1:8080" # 这个是机器上v2ray开放的grpc地址
```

### 配置v2ray:

> 这只是一份参考的配置，
> 关键的部分在于`stats/api/policy/routing`
> 另外如果需要对接 django-sspanel的话，配置里的inbound的`port/tag`必须和面板后台里的配置相同

```json
{
"stats": {},
"api": {
    "services": [
    "HandlerService",
    "StatsService"
    ],
    "tag": "api"
},
"policy": {
    "levels": {
    "0": {
        "handshake": 4,
        "connIdle": 300,
        "uplinkOnly": 2,
        "downlinkOnly": 5,
        "statsUserUplink": true,
        "statsUserDownlink": true,
        "bufferSize": 10240
    }
    },
    "system": {
    "statsInboundUplink": true,
    "statsInboundDownlink": true
    }
},
"inbound": {
    "port": 10086,
    "protocol": "vmess",
    "settings": {
    "clients": []
    },
    "streamSettings": {
    "network": "tcp"
    },
    "tag": "proxy"
},
"inboundDetour": [
    {
    "listen": "127.0.0.1",
    "port": 8080,
    "protocol": "dokodemo-door",
    "settings": {
        "address": "127.0.0.1"
    },
    "tag": "api"
    }
],
"log": {
    "loglevel": "debug"
},
"outbound": {
    "protocol": "freedom",
    "settings": {}
},
"routing": {
    "settings": {
    "rules": [
        {
        "inboundTag": [
            "api"
        ],
        "outboundTag": "api",
        "type": "field"
        }
    ]
    },
    "strategy": "rules"
}
}
```