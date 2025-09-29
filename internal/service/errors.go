package service

import "errors"

// 定义服务层错误
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrRoomNotFound       = errors.New("room not found")
	ErrNotRoomMember      = errors.New("not a room member")
	ErrMessageNotFound    = errors.New("message not found")
	ErrInvalidOperation   = errors.New("invalid operation")
)
