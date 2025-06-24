# Speedtest
测速工具

## 介绍
内置节点来自广东移动,支持自定义节点添加  
该项目为实验性项目,不保证正常使用  

## 自定义节点Json格式
```
{
    "referenceApacheAgents": [
        {
            "bandwidth": 1250000000,
            "blockSize": 0,
            "description": "Cloudflare",
            "hostIp": "speed.cloudflare.com:443",
            "location": 0,
            "name":"Cloudflare",
            "operator": 100000,
            "protocol": "https",
            "downloadpath":"__down?bytes=25000000",
            "uploadpath":"__down"
        }
    ]
}
```
放于``local``文件夹内  

## License
GPL-3.0 license
