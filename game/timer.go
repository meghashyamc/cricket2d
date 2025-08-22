package game

import "time"

type Timer struct {
	currentTime time.Duration
	targetTime  time.Duration
}

func NewTimer(target time.Duration) *Timer {
	return &Timer{
		currentTime: 0,
		targetTime:  target,
	}
}

func (t *Timer) Update() {
	t.currentTime += time.Second / 60 // 60 FPS
}

func (t *Timer) IsReady() bool {
	return t.currentTime >= t.targetTime
}

func (t *Timer) Reset() {
	t.currentTime = 0
}