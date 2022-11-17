package network

import (
	"fmt"
)

const (
	intLen int = 64
)

type BitMap struct {
	size int
	mp   []int64
}

func NewBitMap(size int) *BitMap {
	ret := BitMap{
		size: size,
	}
	num := size/intLen + 1
	ret.mp = make([]int64, num)
	for i := 0; i < num; i++ {
		ret.mp[i] = 0
	}
	return nil
}

func (bm *BitMap) Set(pos int) error {
	if pos > bm.size {
		return fmt.Errorf("invalid position")
	}
	index := pos / intLen
	bitPos := pos % intLen
	bm.mp[index] = bm.mp[index] | (1 << bitPos)
	return nil
}

func (bm *BitMap) Unset(pos int) error {
	if pos > bm.size {
		return fmt.Errorf("invalid position")
	}
	index := pos / intLen
	bitPos := pos % intLen
	bm.mp[index] = bm.mp[index] & (^(1 << bitPos))
	return nil
}

func (bm *BitMap) IsSet(pos int) (bool, error) {
	if pos > bm.size {
		return false, fmt.Errorf("invaild position")
	}
	index := pos / intLen
	bitPos := pos % intLen
	if bm.mp[index]&(1<<bitPos) != 0 {
		return true, nil
	}
	return false, nil
}
