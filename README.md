<p align="center"><img src="https://testingcf.jsdelivr.net/gh/4444TENSEI/CDN/img/avatar/AngelDog/AngelDog-rounded.png" alt="Logo"
    width="200" height="200"/></p>
<h1 align="center">EndpointJoiner</h1>
<h3 align="center">HTTP端点爆破(测试工具)</h3>
<p align="center">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
    <img src="https://img.shields.io/badge/json-5E5C5C?style=for-the-badge&logo=json&logoColor=white" />
</p>


<br/>



<hr/>

# 部署

### 拉取项目

```
git clone https://github.com/4444TENSEI/EndpointJoiner
```

### 启动

```
go run main.go
```



<hr/>

# 修改配置文件

### `config.json`示例:

```json
{
    "constituentWord": [
        "song",
        "songs",
        "rank",
        "user"
    ],
    "endWord": [
        "/",
        "/666666",
        "/rank?id=666666"
    ],
    "mutexGroups": {
        "song": [
            "songs"
        ]
    },
    "baseDomains": [
        "https://xxx.xxx.com/api",
        "https://xxx.xxx.com/api/v1",
        "https://xxx.xxx.com/api/v6"
    ],
    "concurrency": 1000
}
```

# 参数说明

|       参数       | 值                                                           |
| :--------------: | :----------------------------------------------------------- |
|      constituentWord      | `字符串`，多个随机值，生成时自动算出所有可能的组合顺序 |
|   endWord   | `字符串`，必定生成在尾部的自定义部分 |
| mutexGroups | `键值对`，定义冲突词，键名与值不会同时出现在最后生成的随机端点中。 |
|     baseDomains     | `字符串`，根域名，尾部不要带``/``斜杠 |
| concurrency | `数值`，并发数，越大越快 |

