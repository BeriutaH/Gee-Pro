package geecache

// PeerPicker 是必须实现的接口，用于查找拥有特定密钥的对等方
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 方法用于根据传入的 key 选择相应节点 PeerGetter
}

// PeerGetter 是 Peer 必须实现的接口
type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 从对应 group 查找缓存值
}
