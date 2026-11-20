package xkit

import "fmt"

// BiMap[K, V] 双向映射
type BiMap[K, V comparable] struct {
	k2v map[K]V
	v2k map[V]K
}

// NewBiMap 创建空的双向映射
func NewBiMap[K, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{
		k2v: make(map[K]V),
		v2k: make(map[V]K),
	}
}

// NewBiMapKV 从已有 map 创建双向映射
func NewBiMapKV[K, V comparable](kv map[K]V) *BiMap[K, V] {
	m := NewBiMap[K, V]()
	for k, v := range kv {
		m.AddKV(k, v)
	}
	return m
}

// AddKV 添加一个键值对；k 或 v 已存在时先移除旧映射，保持一一对应
func (m *BiMap[K, V]) AddKV(k K, v V) {
	if old, ok := m.k2v[k]; ok {
		delete(m.v2k, old)
	}
	if old, ok := m.v2k[v]; ok {
		delete(m.k2v, old)
	}
	m.k2v[k] = v
	m.v2k[v] = k
}

// MustK2V 按 key 查找 value，不存在则 panic
func (m *BiMap[K, V]) MustK2V(k K) V {
	v, ok := m.k2v[k]
	if !ok {
		panic(fmt.Sprintf("xkit.BiMap: key %v not found", k))
	}
	return v
}

// MustV2K 按 value 查找 key，不存在则 panic
func (m *BiMap[K, V]) MustV2K(v V) K {
	k, ok := m.v2k[v]
	if !ok {
		panic(fmt.Sprintf("xkit.BiMap: value %v not found", v))
	}
	return k
}

// K2V 按 key 查找 value
func (m *BiMap[K, V]) K2V(k K) (V, bool) {
	v, ok := m.k2v[k]
	return v, ok
}

// V2K 按 value 查找 key
func (m *BiMap[K, V]) V2K(v V) (K, bool) {
	k, ok := m.v2k[v]
	return k, ok
}

// KExist 检查 key 是否存在
func (m *BiMap[K, V]) KExist(k K) bool {
	_, ok := m.k2v[k]
	return ok
}

// VExist 检查 value 是否存在
func (m *BiMap[K, V]) VExist(v V) bool {
	_, ok := m.v2k[v]
	return ok
}

// Ks 返回所有 key
func (m *BiMap[K, V]) Ks() []K {
	keys := make([]K, 0, len(m.k2v))
	for k := range m.k2v {
		keys = append(keys, k)
	}
	return keys
}

// Vs 返回所有 value
func (m *BiMap[K, V]) Vs() []V {
	vals := make([]V, 0, len(m.v2k))
	for v := range m.v2k {
		vals = append(vals, v)
	}
	return vals
}
