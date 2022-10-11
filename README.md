# KubeFinder

![logo](docs/logo.png)

一个独立的容器文件服务器, 可以通过该项目查看容器内(namespace/pod/container)的文件目录结构/列表, 下载文件或者上传文件到指定目录中. 使用golang开发, 直接操作主机 `/proc/<pid>/root` 目录, 速度很快.

这是一个后端项目, **仅提供API**, 前端对接之后可以是这样的

![fileserver1](docs/fileserver1.png)

也可以是这样的：

![fileserver2](docs/fileserver2.png)

项目名取自MacOS Finder, 希望它像Finder一样好用.

目前该项目处在`alpha`阶段.

## 为什么做这个项目?

- 没发现开源好用的基于kubernetes的容器文件服务项目
- 我们本身的需求: 
  - 一些可视化项目需要获取容器内文件结构
  - 有不少上传文件到容器中的需求
- `kubectl cp`中的实现
  - 仅实现了上传和下载, 没有文件结构
  - 通过exec逻辑实现相对复杂, 而且依赖`tar`命令, 社区已经在[讨论](https://github.com/kubernetes/kubernetes/issues/58512)更改实现方式

## 接口

### 通用说明

接口错误都是以json格式返回error字段说明，http状态码根据错误情况不同一般为：`404`、`500`, 比如：

1. token 过期

```
Status Code: 403
{
    "error": "token is expired by 14h5m25s"
}
```

2. 未提供 token

```
Status Code: 403
{
    "error": "token not found"
}
```

主要包括获取token接口, 列表和上传接口, 首先获取接口认证用的jwt, 然后使用文件列表和文件上传接口. 下面`接口2`和`接口3`已经配置允许cors请求

### 1 获取token接口

获取用于后续接口调用的jwt, 有效期默认30min

#### URL

`GET /apis/v1/auth/token`

#### Header

使用BasicAuth认证, 账号密码在config.yaml文件中配置, 请参考config.example.yaml中clients部分的配置:

```shell
username: <username>
password: <password>
```

#### 响应

响应码200，同时返回token。

```
Status Code: 200
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJyb25ncWl5dW4iLCJleHAiOjE2NTg3OTk3MjUsImlzcyI6Imt1YmUtZmlsZXNlcnZlciJ9.DA1CvAdSDs_p3c3BjQpvHX0s4UjNj4mLmP4mkZqYlro"
}
```

### 2 文件列表/下载接口

获取集群中某一个容器的文件列表

#### URL

`GET /apis/v1/namespaces/<namespace>/pods/<pod>/containers/<container>/files?subpath=<subpath>&token=<jwt>`

#### URL参数说明

|param|type|remark|
|-|-|-|
|namespace|string|容器所在名称空间|
|pod|string|容器所在pod|
|container|string|容器|

#### Query string说明

|item|type|remark|
|-|-|-|
|subpath|string|starts with `/`, ends with file or folder name, you **SHOULD NOT** end with `/ or .`|

subpath, for example:

|subpath|type|remark|
|-|-|-|
|/|string|root path|
|/bin|string|directory|
|/home/jimy/app.jar|string|file|

#### 响应

`http status: 200 OK`

```
{
    "subpath": "/home",
    "files": [
        {
            "name": "core",
            "size": 11,
            "mode": "Lrwxrwxrwx",
            "modtime": "2022-06-13T07:12:24.612837757Z",
            "isdir": false,
            "groupname": "root",
            "username": "root",
            "link": "/proc/kcore"
        },
        {
            "name": "curl_user",
            "size": 26,
            "mode": "dgrwxr-xr-x",
            "modtime": "2022-07-26T00:57:50.46762204Z",
            "isdir": true,
            "groupname": "curl_group",
            "username": "curl_user",
            "link": ""
        }
    ]
}

```

说明:

|item|type|remark|
|-|-|-|
|name|string|文件/文件夹名称|
|size|int|大小,单位/字节B|
|mode|string|权限|
|modtime|time|最后修改时间|
|isdir|bool|是否文件夹, 如果不是文件夹, 则可以下载|
|groupname|string|文件所属组名|
|username|string|文件所属用户名|
|link|string|符号连接|

### 3 文件上传接口

上传文件到指定容器的路径(subpath)下, **路径必须存在**. 上传之后的文件权限: `0777`

`POST /apis/v1/namespaces/<namespace>/pods/<pod>/containers/<container>/files?subpath=<subpath>&token=<jwt>`

#### query string

说明同上, subpath 必须是文件夹

#### header

```
Content-Type: multipart/form-data
```

#### body

|param|type|remark|
|-|-|-|
|file|file|file to upload, single file|

#### response

```
Status Code: 200
{
    "message": "file /host/proc/9698/root/home/curl_user/test_dir/LICENSE uploaded successfully"
}
```

## 架构图

- 整体架构

![pasted-image](docs/arch-1.png)

- 组件

![pasted-image](docs/arch-2.png)

- 数据流图

![pasted-image](docs/arch-3.png)

## 测试

实际传输速度取决于网络环境。以下数据为内网测试， 仅供参考：

1. 大文件上传 传输速度130Mi/s， 文件大小 918Mi 用时 48.68s

![upload](docs/upload.png)

2. 大文件下载 传输速度 10Mi/s， 文件大小 918Mi 用时 60s

![download.png](docs/download.png)

## 部署

默认配置在 `deploy/config.prod.yaml` 中， 如果需要变更配置。使用configmap挂载到容器中`/etc/kube-finder/config.yaml`就可以了。 

```
# 部署/变更部署
kubectl -n <ns> apply -f deploy/daemonset.yaml
```
