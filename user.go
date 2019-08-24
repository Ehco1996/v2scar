package v2scar

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// User V2ray User
type User struct {
	Email           string `json:"email"`
	UUID            string `json:"uuid"`
	AlterId         uint32 `json:"alter_id"`
	Level           uint32 `json:"level"`
	Enable          bool   `json:"enable"`
	UploadTraffic   int64  `json:"upload_traffic"`
	DownloadTraffic int64  `json:"download_traffic"`
	running         bool
}

func newUser(email, uuid string, level, alterId uint32, enable bool) *User {
	return &User{
		Email:   email,
		UUID:    uuid,
		Level:   level,
		Enable:  enable,
		AlterId: alterId,
	}
}

func (u *User) setUploadTraffic(ut int64) {
	atomic.StoreInt64(&u.UploadTraffic, ut)
}

func (u *User) setDownloadTraffic(dt int64) {
	atomic.StoreInt64(&u.DownloadTraffic, dt)
}

func (u *User) setEnable(enable bool) {
	// NOTE not thread safe!
	u.Enable = enable
}

func (u *User) setRunning(status bool) {
	// NOTE not thread safe!
	u.running = status
}

// UserPool user pool
type UserPool struct {
	access sync.RWMutex
	users  map[string]*User
}

// NewUserPool New UserPool
func NewUserPool() *UserPool {
	// map key : email
	return &UserPool{
		users: make(map[string]*User),
	}
}

// CreateUser get create user
func (up *UserPool) CreateUser(email, uuid string, level, alterId uint32, enable bool) (*User, error) {
	up.access.Lock()
	defer up.access.Unlock()

	if user, found := up.users[email]; found {
		return user, errors.New(fmt.Sprintf("User Already Exists Email: %s", email))
	} else {
		user := newUser(email, uuid, level, alterId, enable)
		up.users[user.Email] = user
		return user, nil
	}
}

// GetUserByEmail get user by email
func (up *UserPool) GetUserByEmail(email string) (*User, error) {
	up.access.Lock()
	defer up.access.Unlock()

	if user, found := up.users[email]; found {
		return user, nil
	} else {
		return nil, errors.New(fmt.Sprintf("User Not Found Email: %s", email))
	}
}

// GetAllUsers GetAllUsers
func (up *UserPool) GetAllUsers() []*User {
	up.access.Lock()
	defer up.access.Unlock()

	users := make([]*User, 0, len(up.users))
	for _, user := range up.users {
		users = append(users, user)
	}
	return users
}
