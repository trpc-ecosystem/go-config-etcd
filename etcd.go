//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package etcd

import (
	"context"
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"

	"trpc.group/trpc-go/trpc-go/config"
	"trpc.group/trpc-go/trpc-go/plugin"
)

func init() {
	plugin.Register(pluginName, NewPlugin())
}

const (
	pluginName = "etcd"
	pluginType = "config"
)

// ErrNotImplemented not implemented error
var ErrNotImplemented = errors.New("not implemented")

// NewPlugin initializes the plugin.
func NewPlugin() plugin.Factory {
	return &etcdPlugin{}
}

// etcdPlugin etcd Configuration center plugin.
type etcdPlugin struct{}

// Type implements plugin.Factory.
func (p *etcdPlugin) Type() string {
	return pluginType
}

// Setup implements plugin.Factory.
func (p *etcdPlugin) Setup(name string, decoder plugin.Decoder) error {
	cfg := clientv3.Config{}
	err := decoder.Decode(&cfg)
	if err != nil {
		return err
	}
	c, err := New(cfg)
	if err != nil {
		return err
	}
	config.SetGlobalKV(c)
	config.Register(c)
	return nil
}

// Client etcd client.
type Client struct {
	cli *clientv3.Client
}

// New creates an etcd client instance.
func New(cfg clientv3.Config) (*Client, error) {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

// watchResponse represents response of etcd watch command.
type watchResponse struct {
	val       string
	md        map[string]string
	eventType config.EventType
}

// Value implements the config.Response interface.
func (r *watchResponse) Value() string {
	return r.val
}

// Event implements the config.Response interface.
func (r *watchResponse) Event() config.EventType {
	return r.eventType
}

// MetaData implements the config.Response interface.
func (r *watchResponse) MetaData() map[string]string {
	return r.md
}

// getResponse represents response of etcd get command.
type getResponse struct {
	val string
	md  map[string]string
}

// Value implements the config.Response interface.
func (r *getResponse) Value() string {
	return r.val
}

// Event implements the config.Response interface.
func (r *getResponse) Event() config.EventType {
	return config.EventTypeNull
}

// MetaData implements the config.Response interface.
func (r *getResponse) MetaData() map[string]string {
	return r.md
}

// Get Obtains the configuration content value according to the key, and implement the config.KV interface.
func (c *Client) Get(ctx context.Context, key string, _ ...config.Option) (config.Response, error) {
	result, err := c.cli.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	rsp := &getResponse{
		md: make(map[string]string),
	}

	if result.Count > 1 {
		// TODO: support multi keyvalues
		return nil, ErrNotImplemented
	}

	for _, v := range result.Kvs {
		rsp.val = string(v.Value)
	}
	return rsp, nil
}

// Put creates or updates the configuration content value to implement the config.KV interface.
func (c *Client) Put(ctx context.Context, key, val string, opts ...config.Option) error {
	return ErrNotImplemented
}

// Del deletes the configuration item key and implement the config.KV interface.
func (c *Client) Del(ctx context.Context, key string, opts ...config.Option) error {
	return ErrNotImplemented
}

// Watch monitors configuration changes and implements the config.Watcher interface.
func (c *Client) Watch(ctx context.Context, key string, opts ...config.Option) (<-chan config.Response, error) {
	rspCh := make(chan config.Response, 1)
	go c.watch(ctx, key, rspCh)
	return rspCh, nil
}

// Name returns plugin name.
func (c *Client) Name() string {
	return pluginName
}

// watch adds watcher for etcd changes.
func (c *Client) watch(ctx context.Context, key string, rspCh chan config.Response) {
	rch := c.cli.Watch(ctx, key)
	for r := range rch {
		for _, ev := range r.Events {
			rsp := &watchResponse{
				val:       string(ev.Kv.Value),
				md:        make(map[string]string),
				eventType: config.EventTypeNull,
			}
			switch ev.Type {
			case clientv3.EventTypePut:
				rsp.eventType = config.EventTypePut
			case clientv3.EventTypeDelete:
				rsp.eventType = config.EventTypeDel
			default:
			}
			rspCh <- rsp
		}
	}
}
