package v2scar

import (
	"sync"
	"sync/atomic"
)

// User V2ray User
type User struct {
	Email           string `json:"_"`
	UploadTraffic   int64  `json:"upload_traffic"`
	DownloadTraffic int64  `json:"download_traffic"`
}

func (u *User) setUploadTraffic(ut int64) {
	atomic.StoreInt64(&u.UploadTraffic, ut)
}

func (u *User) setDownloadTraffic(dt int64) {
	atomic.StoreInt64(&u.DownloadTraffic, dt)
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

// GetOrCreateUser GetOrCreateUser
func (up *UserPool) GetOrCreateUser(email string) *User {
	up.access.Lock()
	defer up.access.Unlock()

	if user, found := up.users[email]; found {
		return user
	}
	user := &User{
		Email: email,
	}
	up.users[user.Email] = user
	return user
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
