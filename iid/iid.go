package iid

import (
	"sync"
	"time"
)

const (
	twepoch = int64(1601481600000)
	aidBits = uint(5)
	nidBits = uint(9)
	sBits   = uint(10)

	/*
	 * 1 符号位  |  39 时间戳                                     | 5 区域  |  9 节点       | 10 （毫秒内）自增ID
	 * 0        |  0000000 00000000 00000000 00000000 00000000  | 00000  | 000000 000   |  000000 0000
	 *
	 */

	nidShift = sBits //左移次数
	aidShift = sBits + nidBits
	tShift   = sBits + aidBits + nidBits
	sMask    = -1 ^ (-1 << sBits)
	nidMask  = -1 ^ (-1 << nidBits)
	aidMask  = -1 ^ (-1 << aidBits)
)

const TWEPOCH = twepoch

func Milliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type IID struct {
	s    int64 //序号
	aid  int64 //区域ID
	nid  int64 //节点ID
	lock sync.Mutex
}

type IIDComponent struct {
	AID          int64
	NID          int64
	Milliseconds int64
	S            int64
}

func NewIID(aid int64, nid int64) *IID {
	v := IID{}
	v.s = 0
	v.nid = nid
	v.aid = aid
	v.lock = sync.Mutex{}
	return &v
}

func (v *IID) NewID() int64 {
	v.lock.Lock()
	defer v.lock.Unlock()

	id := Milliseconds()
	v.s = (v.s + 1) & sMask

	return ((id - twepoch) << tShift) | (v.aid << aidShift) | (v.nid << nidShift) | v.s
}

func GetComponent(id int64) *IIDComponent {
	return &IIDComponent{S: id & sMask,
		NID:          (id >> nidShift) & nidMask,
		AID:          (id >> aidShift) & aidMask,
		Milliseconds: (id >> tShift) + twepoch}
}

func GetID(v *IIDComponent) int64 {
	return ((v.Milliseconds - twepoch) << tShift) | (v.AID << aidShift) | (v.NID << nidShift) | v.S
}
