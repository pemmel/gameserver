package server

import (
	"crypto/aes"
	"crypto/cipher"
	"sync"
	"time"
	"unsafe"

	"github.com/bytedance/gopkg/lang/fastrand"
)

// shared represents a shared instance of SessionContainer for managing sessions.
var shared *SessionContainer

// init initializes the shared SessionContainer with a capacity of cap.
func init() {
	const cap int = 1e3
	shared = &SessionContainer{
		data:  make([]*Session, 0, cap),
		empty: 0,
	}
}

// SharedSession returns the shared instance of SessionContainer.
func SharedSession() *SessionContainer {
	return shared
}

// GenerateKey generates a random key for session encryption.
func GenerateKey(b []byte) {
	n := len(b) / 4
	r := len(b) % 4

	for i := 0; i < n; i++ {
		ptr := (*uint32)(unsafe.Pointer(&b[i*4]))
		*ptr = fastrand.Uint32()
	}

	if r != 0 {
		var tmp [4]byte
		ptr := (*uint32)(unsafe.Pointer(&tmp[0]))
		*ptr = fastrand.Uint32()
		copy(b[n*4:], tmp[:r])
	}
}

// SessionContainer represents a container for managing sessions.
type SessionContainer struct {
	mutex sync.Mutex
	data  []*Session
	empty int
}

// NewSession creates a new session with the provided NewSession function and
// adds it to the session container only if able to reserve new session index
// (sidx) and able to create a new session via NewSession function.
//
// Parameters:
//   - new: A function for creating a new session.
//   - uid: The user ID associated with the new session.
//
// Returns:
//   - *Session: A pointer to the newly created session, or nil if the session
//     index reservation fails or if the NewSession function fails
//     to create a new instance.
func (c *SessionContainer) NewSession(new NewSession, uid uint) *Session {
	var sidx uint32
	if !c.reserveSidx(&sidx) {
		return nil
	}
	return c.register(new, uid, sidx)
}

// OutOfBounds checks if the provided session index is out of bounds, i.e., if it
// exceeds the length of the session container's data slice.
//
// Parameters:
//   - sidx: The session index to check.
//
// Returns:
//   - bool: True if the session index is out of bounds, false otherwise..
func (c *SessionContainer) OutOfBounds(sidx uint32) bool {
	return int64(sidx) >= int64(len(c.data))
}

// CanGrow checks if the session container can grow by comparing its current length
// with the maximum value that Session.Sidx can represent.
//
// Returns:
//   - bool: True if the session container can grow, false otherwise.
func (c *SessionContainer) CanGrow() bool {
	const max = ^uint32(0)
	return int64(len(c.data)) < int64(max)
}

// Remove removes a session from the session container at the specified index.
//
// Parameters:
//   - sidx: The session index of the session to remove.
//
// Returns:
//   - *Session: A pointer to the removed session, or nil if the session index
//     is out of bounds.
func (c *SessionContainer) Remove(sidx uint32) *Session {
	if c.OutOfBounds(sidx) {
		return nil
	}
	c.mutex.Lock()
	d := c.data[sidx]
	if d != nil {
		c.data[sidx] = nil
		c.empty++
	}
	c.mutex.Unlock()
	return d
}

// Get retrieves a session from the session container based on the provided session index.
//
// Parameters:
//   - sidx: The session index of the session to retrieve.
//
// Returns:
//   - *Session: A pointer to the session at the specified index, or nil if the index is
//     out of bounds.
func (c *SessionContainer) Get(sidx uint32) *Session {
	if c.OutOfBounds(sidx) {
		return nil
	}
	return c.data[sidx]
}

// GetFromUid retrieves a session from the session container based on the provided user ID.
//
// Parameters:
//   - uid: The user ID associated with the session to retrieve.
//
// Returns:
//   - *Session: A pointer to the session with the provided user ID, or nil if no session
//     with the user ID is found.
func (c *SessionContainer) GetFromUid(uid uint) *Session {
	for _, d := range c.data {
		if d != nil && d.Uid == uid {
			return d
		}
	}
	return nil
}

// findEmptySpace searches the session container for an empty slot.
//
// Returns:
//   - int: The index of the first empty slot in the session container,
//     or -1 if no empty slots are found.
func (c *SessionContainer) findEmptySpace() int {
	for i, d := range c.data {
		if d == nil {
			return i
		}
	}
	return -1
}

// register registers a new session in the session container at the specified index.
// If the session creation using the provided NewSession function fails, a nil pointer
// is stored in the container to mark the slot as empty. The session index and the
// session itself are updated in the container upon successful registration.
//
// Parameters:
//   - new: A function for creating a new session.
//   - uid: The user ID associated with the new session.
//   - sidx: The index where the new session will be registered, obtained via the
//     reserveSidx function to ensure a valid and available index.
//
// Returns:
//   - *Session: A pointer to the newly registered session, which is identical to
//     the session created by the NewSession function, which could be nil if the session
//     creation fails.
func (c *SessionContainer) register(new NewSession, uid uint, sidx uint32) *Session {
	s := new(uid)
	if s == nil {
		c.data[sidx] = nil
		c.mutex.Lock()
		c.empty++
		c.mutex.Unlock()
	} else {
		s.Sidx = sidx
		c.data[sidx] = s
	}
	return s
}

// reserveSidx reserves a session index for a new session in the session container.
// Upon successful reservation, the provided sidx holds an index to the container
// filled with a filler session. This ensures that sidx will not read as an empty
// space, yet it's not a valid Session pointer.
//
// Parameters:
//   - sidx: A pointer to a uint32 variable where the reserved session index will be stored.
//
// Returns:
//   - bool: True if the session index was successfully reserved, false otherwise.
func (c *SessionContainer) reserveSidx(sidx *uint32) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.empty != 0 {
		*sidx = uint32(c.findEmptySpace())
		c.data[*sidx] = c.filler()
		c.empty--
		return true
	}

	if c.CanGrow() {
		*sidx = uint32(len(c.data))
		c.data = append(c.data, c.filler())
		return true
	}

	return false
}

// filler returns a filler session for uninitialized session slots which value
// is not a nil pointer. This ensures that the findEmptySpace function will not
// assume it as an empty space when searching for available slots.
//
// Returns:
//   - *Session: A filler session for uninitialized session slots.
//
// #nosec
func (c *SessionContainer) filler() *Session {
	return (*Session)(unsafe.Pointer(uintptr(1)))
}

type Session struct {
	Version   uint8
	Sidx      uint32
	Uid       uint
	StateIdx  int
	GameState int
	LoginTime time.Time
	SharedKey [32]byte
	Cipher    cipher.AEAD
	Mutex     sync.Mutex
}

type NewSession = func(uid uint) *Session

// NewSessionV1 creates a new session with version 1 format, generating a unique
// key and initializing the cipher for encryption. If any error occurs during
// the key generation or cipher initialization, it returns nil.
//
// Parameters:
//   - uid: The user ID associated with the new session.
//
// Returns:
//   - *Session: A pointer to the newly created session with version 1 format,
//     or nil if an error occurs during key generation or cipher initialization.
func NewSessionV1(uid uint) *Session {
	var sk [32]byte
	GenerateKey(sk[:])

	cb, err := aes.NewCipher(sk[:])
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(cb)
	if err != nil {
		return nil
	}

	return &Session{
		Version:   1,
		Uid:       uid,
		StateIdx:  -1,
		GameState: GameState_Idle,
		SharedKey: sk,
		Cipher:    gcm,
		LoginTime: time.Now(),
	}
}
