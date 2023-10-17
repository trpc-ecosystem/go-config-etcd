//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
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
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"trpc.group/trpc-go/trpc-go/config"
)

func TestNew(t *testing.T) {
	t.Run("new_client_fail_case", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(clientv3.New, nil, fmt.Errorf("construct client fail"))
		defer patches.Reset()
		_, err := New(clientv3.Config{})
		assert.NotNil(t, err)
	})
}

func TestClient_Del(t *testing.T) {
	c := &Client{}
	err := c.Del(context.Background(), "test_key")
	assert.NotNil(t, err)
}

func TestClient_Put(t *testing.T) {
	c := &Client{}
	err := c.Put(context.Background(), "test_key", "test_val")
	assert.NotNil(t, err)
}

func TestClient_Get(t *testing.T) {
	mKV := &mockKV{}
	mockETCDClient := &clientv3.Client{
		KV: mKV,
	}
	c := &Client{cli: mockETCDClient}
	testKey := "test_key"
	testVal := "test_val"

	t.Run("right_case", func(t *testing.T) {
		mKV.setGetRsp(&clientv3.GetResponse{
			Count: 1,
			Kvs: []*mvccpb.KeyValue{
				{
					Key:   []byte(testKey),
					Value: []byte(testVal),
				},
			},
		}, nil)
		rsp, err := c.Get(context.Background(), testKey)
		assert.Nil(t, err)
		assert.Equal(t, rsp.Value(), testVal)
		// metaData没有实现，结果为空
		assert.Equal(t, rsp.MetaData(), map[string]string{})
		assert.Equal(t, rsp.Event(), config.EventTypeNull)
	})

	t.Run("get_err_case", func(t *testing.T) {
		mKV.setGetRsp(nil, fmt.Errorf("get from etcd fail"))
		_, err := c.Get(context.Background(), testKey)
		assert.NotNil(t, err)
	})

	t.Run("cnt_over_case", func(t *testing.T) {
		mKV.setGetRsp(&clientv3.GetResponse{
			Count: 2,
			Kvs: []*mvccpb.KeyValue{
				{
					Key:   []byte(testKey),
					Value: []byte(testVal),
				},
				{
					Key:   []byte(testKey),
					Value: []byte(testVal),
				},
			},
		}, nil)
		_, err := c.Get(context.Background(), testKey)
		assert.NotNil(t, err)
	})
}

func TestClient_Watch(t *testing.T) {
	mw := &mockWatch{}
	mockETCDClient := &clientv3.Client{
		Watcher: mw,
	}
	c := &Client{cli: mockETCDClient}

	f := func() clientv3.WatchChan {
		testChan := make(chan clientv3.WatchResponse)
		go func() {
			defer close(testChan)
			testChan <- clientv3.WatchResponse{
				Events: []*clientv3.Event{
					{
						Type: clientv3.EventTypeDelete,
						Kv: &mvccpb.KeyValue{
							Value: []byte("a"),
						},
					},
					{
						Type: clientv3.EventTypePut,
						Kv: &mvccpb.KeyValue{
							Value: []byte("b"),
						},
					},
				},
			}
		}()
		return testChan
	}

	mw.setWatchChan(f())
	_, err := c.Watch(context.Background(), "a")
	assert.Nil(t, err)
}

func TestEtcdPlugin_Type(t *testing.T) {
	a := &etcdPlugin{}
	assert.Equal(t, a.Type(), pluginType)
}

func TestEtcdPlugin_Setup(t *testing.T) {
	a := &etcdPlugin{}

	t.Run("right_case", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(New, &Client{}, nil)
		defer patches.Reset()
		err := a.Setup("a", &mockPluginDecoder{err: nil})
		assert.Nil(t, err)
	})

	t.Run("decode_fail_case", func(t *testing.T) {
		err := a.Setup("a", &mockPluginDecoder{err: fmt.Errorf("decode fail")})
		assert.NotNil(t, err)
	})

	t.Run("new_fail_case", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(New, nil, fmt.Errorf("new client fail"))
		defer patches.Reset()
		err := a.Setup("a", &mockPluginDecoder{err: nil})
		assert.NotNil(t, err)
	})
}

type mockKV struct {
	getRes *clientv3.GetResponse
	getErr error
}

func (m *mockKV) setGetRsp(rsp *clientv3.GetResponse, err error) {
	m.getRes = rsp
	m.getErr = err
}

func (m *mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return m.getRes, m.getErr
}

func (m *mockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return &clientv3.PutResponse{}, nil
}

func (m *mockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, nil
}

func (m *mockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return &clientv3.CompactResponse{}, nil
}

func (m *mockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (m *mockKV) Txn(ctx context.Context) clientv3.Txn {
	return nil
}

type mockWatch struct {
	c clientv3.WatchChan
}

func (m *mockWatch) setWatchChan(c clientv3.WatchChan) {
	m.c = c
}

func (m *mockWatch) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return m.c
}

func (m *mockWatch) RequestProgress(ctx context.Context) error {
	return nil
}

func (m *mockWatch) Close() error {
	return nil
}

type mockPluginDecoder struct {
	err error
}

func (m *mockPluginDecoder) Decode(conf interface{}) error {
	return m.err
}
