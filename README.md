English | [中文](README_CN.md)

# tRPC-Go etcd configuration plugin

[![Go Reference](https://pkg.go.dev/badge/github.com/trpc-ecosystem/go-config-etcd.svg)](https://pkg.go.dev/github.com/trpc-ecosystem/go-config-etcd)
[![Go Report Card](https://goreportcard.com/badge/trpc.group/trpc-go/trpc-config-etcd)](https://goreportcard.com/report/trpc.group/trpc-go/trpc-config-etcd)
[![LICENSE](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://github.com/trpc-ecosystem/go-config-etcd/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/release/trpc-ecosystem/go-config-etcd.svg?style=flat-square)](https://github.com/trpc-ecosystem/go-config-etcd/releases)
[![Tests](https://github.com/trpc-ecosystem/go-config-etcd/actions/workflows/prc.yml/badge.svg)](https://github.com/trpc-ecosystem/go-config-etcd/actions/workflows/prc.yml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-config-etcd/branch/main/graph/badge.svg)](https://app.codecov.io/gh/trpc-ecosystem/go-config-etcd/tree/main)

The plugin encapsulates [etcd-client](https://github.com/etcd-io/etcd/tree/main/client/v3), facilitating rapid access to configurations in etcd within the tRPC-Go framework.

## Get started

### Step 1

Anonymous import this package

```go
import _ "trpc.group/trpc-go/trpc-config-etcd"
```

### Step 2

In the trpc_go.yaml configuration file, set the Endpoint and Dialtimeout, for the complete configuration, refer to [Config](https://github.com/etcd-io/etcd/blob/client/v3.5.9/client/v3/config.go#L26)

```yaml
plugins:                 
  config:
    etcd:
      endpoints:
        - localhost:2379
      dialtimeout: 5s
```

### Step 3

After calling trpc.NewServer, retrieve the etcd configuration item.

```go
func main() {
	trpc.NewServer()

    // Get the configuration item with the key "foo"
	value, err := config.GetString("foo")
	if err != nil {
		panic(err)
	}
	fmt.Println(value)

      // Watch changes to the configuration item with the key "foo"
	ch, err := config.Get("etcd").Watch(context.Background(), "foo")
	if err != nil {
		panic(err)
	}
	for rsp := range ch {
		fmt.Println(rsp.Value())
	}
}
```

## Notes

The plugin currently only supports reading configurations, and the Get and Put functions are not yet implemented.
