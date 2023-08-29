# tRPC-Go etcd 配置插件

[English](README.md) | 中文

插件封装了 [etcd-client](https://github.com/etcd-io/etcd/tree/main/client/v3)，便于在 tRPC-Go 框架中快速访问 etcd 中的配置。

## 使用方法

### 第一步

匿名 import 此包

```go
import _ "trpc.group/trpc-go/trpc-config-etcd"
```

### 第二步

在 trpc_go.yaml 配置文件中设置 Endpoint 和 Dialtimeout 信息，完整配置见 [Config](https://github.com/etcd-io/etcd/blob/client/v3.5.9/client/v3/config.go#L26)

```yaml
plugins:                 
  config:
    etcd:
      endpoints:
        - localhost:2379
      dialtimeout: 5s
```

### 第三步

在 trpc.NewServer 之后，获取 etcd 配置信息

```go
func main() {
	trpc.NewServer()

    // 获取 key 为 "foo" 的配置项
	value, err := config.GetString("foo")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)

    // 监听 key 为 "foo" 的配置项变化
	ch, err := config.Get("etcd").Watch(context.Background(), "foo")
	if err != nil {
		panic(err)
	}
	for rsp := range ch {
		fmt.Println(rsp.Value())
	}
}
```

## 注意事项

插件暂时只支持读取配置，Get、Put 功能暂未实现。