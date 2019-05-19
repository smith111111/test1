package rpc

import "sync/atomic"

type lb struct {
	target []string
	uint64
}

func NewLB(target []string) *lb {
	return &lb{target: target}
}

// 实现基本轮寻负载均衡
// TODO：服务发现
func (lb *lb) Get() string {
	var old uint64
	for {
		old = atomic.LoadUint64(&lb.uint64)
		if atomic.CompareAndSwapUint64(&lb.uint64, old, old+1) {
			break
		}
	}
	return lb.target[old%uint64(len(lb.target))]
	//return lb.target[0]
}