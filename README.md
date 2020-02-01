# v2scar
Sidecar For V2ray

![go-releaser](https://github.com/Ehco1996/v2scar/workflows/go-releaser/badge.svg)

## 原理介绍

该项目需要以`sidecar`的形式和`v2ray`部署在一台机器上

他通过grpc接口动态添加/删除V2ray的`Vmess`用户,并统计流量

通过接入`django-sspanel`的api，可以动态调节v2ray的Vmess用户

目前实现了以下几个接口：

* AddInboundUser
* RemoveInboundUser
* GetAndResetUserTraffic

## 使用说明

* 可以直接以cli的形式运行:

`./v2scar --api-endpoint="xxx" --grpc-endpoint="127.0.0.1:8080" --sync-time=60`

* 或者通过配置环境变量来运行:

```bash
export V2SCAR_SYNC_TIME=60 # 和django-sspanel同步的时间间隔
export V2SCAR_API_ENDPOINT="xxxx" # 这个是django-sspanel的sync api 地址
export V2SCAR_GRPC_ENDPOINT="127.0.0.1:8080" # 这个是机器上v2ray开放的grpc地址
```

## 配置V2ray:

> 这只是一份参考的配置，
> 关键的部分在于`stats/api/policy/routing`
> 另外如果需要对接 django-sspanel的话，配置里的inbound的`port/tag/level`必须和面板后台里的配置相同

```json
{
    "stats": {},
    "api": {
        "tag": "api",
        "services": [
            "HandlerService",
            "StatsService"
        ]
    },
    "log": {
        "loglevel": "warning"
    },
    "policy": {
        "levels": {
            "0": {
                "statsUserUplink": true,
                "statsUserDownlink": true
            }
        },
        "system": {
            "statsInboundUplink": true,
            "statsInboundDownlink": true
        }
    },
    "inbounds": [
        {
            "tag": "proxy",
            "port": 10086,
            "protocol": "vmess",
            "settings": {
                "clients": []
            }
        },
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
    "outbounds": [
        {
            "protocol": "freedom",
            "settings": {}
        }
    ],
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
