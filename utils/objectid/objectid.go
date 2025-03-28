package objectid

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	// 定义各部分的字节长度
	timestampLen = 4
	machineLen   = 5
	processLen   = 2
	incLen       = 3
)

var (
	// 全局自增序列（需要加锁）
	objectIDCounter uint32
	mutex           sync.Mutex

	// 机器标识（启动时初始化）
	machineID []byte
)

func init() {
	// 初始化机器标识
	initMachineID()
	// 随机初始化计数器
	objectIDCounter = rand.Uint32()
}

// 生成机器标识（5字节）
func initMachineID() {
	machineID = make([]byte, machineLen)

	// 3字节：主机名哈希
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	hash := md5.Sum([]byte(hostname))
	copy(machineID, hash[:3])

	// 2字节：进程ID + 网卡MAC哈希
	pid := os.Getpid()
	machineID[3] = byte(pid >> 8)
	machineID[4] = byte(pid)

	// 如果可能，加入网卡信息
	if interfaces, err := net.Interfaces(); err == nil {
		for _, i := range interfaces {
			if len(i.HardwareAddr) > 0 {
				hash := md5.Sum(i.HardwareAddr)
				machineID[4] ^= hash[0] // 简单混合
				break
			}
		}
	}
}

// 生成新的ObjectId
func New() string {
	var id [12]byte

	// 4字节时间戳（秒级）
	timestamp := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(id[:timestampLen], timestamp)

	// 5字节机器标识
	copy(id[timestampLen:timestampLen+machineLen], machineID)

	// 3字节自增序列（需要加锁）
	mutex.Lock()
	counter := objectIDCounter
	objectIDCounter++
	mutex.Unlock()

	// 写入计数器（取低24位）
	id[9] = byte(counter >> 16)
	id[10] = byte(counter >> 8)
	id[11] = byte(counter)

	return hex.EncodeToString(id[:])
}

// 解析ObjectId
func Parse(id string) (timestamp time.Time, machine string, counter uint32, err error) {
	if len(id) != 24 {
		err = fmt.Errorf("invalid objectid length")
		return
	}

	bytes, err := hex.DecodeString(id)
	if err != nil {
		return
	}

	// 解析时间戳
	secs := binary.BigEndian.Uint32(bytes[:timestampLen])
	timestamp = time.Unix(int64(secs), 0)

	// 解析机器标识
	machine = hex.EncodeToString(bytes[timestampLen : timestampLen+machineLen])

	// 解析计数器
	counter = uint32(bytes[9])<<16 | uint32(bytes[10])<<8 | uint32(bytes[11])

	return
}
