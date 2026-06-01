package services

import "errors"

var ErrNotParticipant = errors.New("user is not a participant of this chat")
