package cluster

import (
	"context"
	"errors"
	"go-redis/resp/client"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type connectionFactory struct {
	Peer string // 表示连接池连接的节点
}

func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	c, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	c.Start()
	return pool.NewPooledObject(c), nil
}

func (f *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	c, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("type mismatch")
	}
	c.Close()
	return nil
}

// 下面这些不需要，返回无作用值即可
func (f *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (f *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (f *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
