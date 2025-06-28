<div align="center">
    <h1>SpeedTest-GD</h1>
</div>
<div align="center">
    <a href="https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/"><img alt="Windows" src="https://img.shields.io/badge/platform-Windows-blue?logo=windowsxp&style=flat-square&color=1E9BFA" /></a>
    <a href="https://github.com/mogumc/MGL/releases"><img alt="Release" src="https://img.shields.io/github/v/release/mogumc/speedtest?logo=visualstudio&style=flat-square&color=1E9BFA"></a>
    <a href="./LICENSE">
        <img alt="GitHub" src="https://img.shields.io/github/license/mogumc/speedtest"/>
    </a>
    <img src="https://komarev.com/ghpvc/?username=mogumc&label=Views&color=orange&style=flat" alt="访问量统计" />
    <h3>SpeedTest-GD</h3>
    <h4>使用Get与Post的测速工具</h4>
</div>

## 介绍
内置了部分节点,不保证可用  
后端采用Get与Post方式测速，只有完整的测试循环才会被记录  
多线程模式下分块越小结果越精确   
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
            "uploadpath":"__up"
        }
    ]
}
```
放于``local``文件夹内  

## Feature 
- [x] 单线程测速
- [x] 多线程测速
- [x] 多节点测速
- [x] 自定义服务器

## License
本项目图标来自`豆包` 版权所有归字节跳动所有  
GPL-3.0 license
