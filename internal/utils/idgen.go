package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// InitIDGenerator 初始化ID生成器
func InitIDGenerator(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// GenerateID 生成唯一ID
func GenerateID() int64 {
	if node == nil {
		// 如果未初始化，使用默认节点ID 1
		_ = InitIDGenerator(1)
	}
	return node.Generate().Int64()
}

// GenerateStringID 生成字符串格式的唯一ID
func GenerateStringID() string {
	return fmt.Sprintf("%d", GenerateID())
}

// GenerateTimeBasedID 生成基于时间的ID
func GenerateTimeBasedID(prefix string) string {
	timestamp := time.Now().UnixNano() / 1000000 // 毫秒级时间戳
	return fmt.Sprintf("%s%d", prefix, timestamp)
}
