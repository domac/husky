package husky

import (
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

//用于生成全局唯一的ID
// Blake Mizerany's `noeqd` https://github.com/bmizerany/noeqd
// Twitter's `snowflake` https://github.com/twitter/snowflake

const (
	nodeIDBits     = uint64(10)
	sequenceBits   = uint64(12)
	nodeIDShift    = sequenceBits
	timestampShift = sequenceBits + nodeIDBits
	sequenceMask   = int64(-1) ^ (int64(-1) << sequenceBits)
	twepoch        = int64(1288834974288)

	MsgIDLength       = 16
	minValidMsgLength = MsgIDLength + 8 + 2 // Timestamp + Attempts
)

var ErrTimeBackwards = errors.New("time has gone backwards")
var ErrSequenceExpired = errors.New("sequence expired")
var ErrIDBackwards = errors.New("ID went backward")

type guid int64

//全局ID生成器
type guidFactory struct {
	sync.Mutex

	nodeID        int64
	sequence      int64
	lastTimestamp int64
	lastID        guid
}

func NewGUIDFactory(nodeID int64) *guidFactory {
	return &guidFactory{
		nodeID: nodeID,
	}
}

func (f *guidFactory) NewGUID() (guid, error) {
	f.Lock()

	// divide by 1048576, giving pseudo-milliseconds
	ts := time.Now().UnixNano() >> 20

	if ts < f.lastTimestamp {
		f.Unlock()
		return 0, ErrTimeBackwards
	}

	if f.lastTimestamp == ts {
		f.sequence = (f.sequence + 1) & sequenceMask
		if f.sequence == 0 {
			f.Unlock()
			return 0, ErrSequenceExpired
		}
	} else {
		f.sequence = 0
	}

	f.lastTimestamp = ts

	id := guid(((ts - twepoch) << timestampShift) |
		(f.nodeID << nodeIDShift) |
		f.sequence)

	if id <= f.lastID {
		f.Unlock()
		return 0, ErrIDBackwards
	}

	f.lastID = id

	f.Unlock()

	return id, nil
}

type MessageID [MsgIDLength]byte

func (g guid) Hex() MessageID {
	var h MessageID
	var b [8]byte

	b[0] = byte(g >> 56)
	b[1] = byte(g >> 48)
	b[2] = byte(g >> 40)
	b[3] = byte(g >> 32)
	b[4] = byte(g >> 24)
	b[5] = byte(g >> 16)
	b[6] = byte(g >> 8)
	b[7] = byte(g)

	hex.Encode(h[:], b[:])
	return h
}
