package storage

type User struct {
	ID       int64  // telegram user id
	Username string //telegram username
	Timezone int    // UTC offset
}
