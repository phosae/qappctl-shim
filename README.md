## 使用方式
Docker
```bash
docker run --rm -d -e ACCESS_KEY=<AK> -e SECRET_KEY=<SK> -p 9100:9100 \
-v /var/run/docker.sock:/var/run/docker.sock zengxu/qappctl-shim:0.2.2
```

直接运行
* 需要预先安装 qappctl
* 如需要镜像上传接口，运行环境必须有 Docker Engine 运行
```
go install github.com/phosae/qappctl-shim@0.2.2

qappctl-shim --access-key <ak> --secret-key <sk>
```

## error
如果发生错误，输入参数错误一律返回 HTTP Status Code 4xx，服务端错误一律返回 5xx

错误信息以 text/plain 格式呈现在 HTTP Body 中

示例如下
```
HTTP/1.1 500 Internal Server Error
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 01 Dec 2022 11:23:45 GMT
Content-Length: 74

err create deploy: exit status 1, 2022/12/01 19:23:45 unknown flag: --id
```

## list images
GET `/images`
```
curl -i localhost:9100/images

HTTP/1.1 200 OK

[
  {
    "name": "nginx",
    "tag": "alpine",
    "ctime": "2022-11-30T18:26:08.935+08:00"
  },
  {
    "name": "ubuntu",
    "tag": "14.04",
    "ctime": "2021-04-22T17:06:34.081+08:00"
  }
]
```

## push image
POST `/images`
```
curl -i -X POST -H 'content-type:application/json' localhost:9100/images -d '{"image": "k8s.gcr.io/pause:3.6"}'

HTTP/1.1 200 OK
```
注意，镜像上传后，前缀会被去除，如 `k8s.gcr.io/pause:3.6` 或者 `k8s.gcr.io/ctr/pause:3.6`，上传后会简化成 `pause:3.6`
## create release
POST `/apps/<app>/releases`
```
curl -X POST -H 'content-type:application/json' localhost:9100/apps/zenx/releases -d '{
    "image": "nginx:alpine",
    "image": "pause:3.6",
    "flavor": "C1M1",
    "port": 80
}'


HTTP/1.1 200 OK

{"name":"v221201-185439"}
```

## list releases
GET `/apps/<app>/releases`
```
curl -i localhost:9100/apps/zenx/releases

HTTP/1.1 200 OK

[
  {
    "name": "v221201-185439",
    "image": "pause:3.6",
    "flavor": "C1M1",
    "port": 80,
    "ctime": "2022-12-01T18:54:40.394+08:00"
  }
]
```
## create deploy
POST `/apps/<app>/deploys`
```
curl -i -X POST -H 'content-type:application/json' localhost:9100/apps/zenx/deploys -d '{
    "release": "v221201-185439",
    "region": "z0",
    "replicas": 1
}'

HTTP/1.1 200 OK

```

## list deploys
GET `/apps/<app>/deploys?region=<region>&release=<release>`

region: must
release: optional

```
curl -i localhost:9100/apps/zenx/deploys?region=z0

HTTP/1.1 200 OK

[
  {
    "id": "g221201-1902-19027-6zvz",
    "release": "v221201-185439",
    "region": "z0",
    "replicas": 1,
    "ctime": "2022-12-01T19:02:19+08:00"
  }
]
```

## delete deploy
DELETE `/apps/<app>/deploys`

id: must
region: must

```
curl -i -X DELETE -H 'content-type:application/json' localhost:9100/apps/zenx/deploys -d '{
    "id": "g221201-1902-19027-6zvz",
    "region": "z0"
}'

HTTP/1.1 200 OK

```

## list intances
GET `/apps/<app>/deploys/<deploy-id>/instances?region=<region>`

```
curl -i localhost:9100/apps/zenx/deploys/c221130-1908-01426-lmkr/instances?region=z0

HTTP/1.1 200 OK

[
  {
    "ctime": "2022-11-30T19:08:01+08:00",
    "id": "c221130-1908-01426-lmkr-64cd4cfd78-25qn5",
    "status": "RUNNING"
  },
  {
    "ctime": "2022-11-30T19:08:01+08:00",
    "id": "c221130-1908-01426-lmkr-64cd4cfd78-nd56z",
    "status": "RUNNING",
    "ips": "101.69.128.12"
  }
]
```
ips 字段表示浮动 IP

