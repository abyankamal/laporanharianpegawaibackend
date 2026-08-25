package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextWeekdayTarget(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	t.Run("Before target time on weekday (Monday 10:00 -> Monday 15:00)", func(t *testing.T) {
		// 2026-03-02 is Monday
		now := time.Date(2026, 3, 2, 10, 0, 0, 0, loc)
		target := nextWeekdayTarget(now, 15, 0, 0)

		assert.Equal(t, time.Monday, target.Weekday())
		assert.Equal(t, 2, target.Day())
		assert.Equal(t, 15, target.Hour())
		assert.Equal(t, 0, target.Minute())
	})

	t.Run("After target time on weekday (Monday 16:00 -> Tuesday 15:00)", func(t *testing.T) {
		// 2026-03-02 is Monday
		now := time.Date(2026, 3, 2, 16, 0, 0, 0, loc)
		target := nextWeekdayTarget(now, 15, 0, 0)

		assert.Equal(t, time.Tuesday, target.Weekday())
		assert.Equal(t, 3, target.Day())
		assert.Equal(t, 15, target.Hour())
	})

	t.Run("After target time on Friday (Friday 16:00 -> Next Monday 15:00)", func(t *testing.T) {
		// 2026-03-06 is Friday
		now := time.Date(2026, 3, 6, 16, 0, 0, 0, loc)
		target := nextWeekdayTarget(now, 15, 0, 0)

		// Next weekday should be Monday (2026-03-09)
		assert.Equal(t, time.Monday, target.Weekday())
		assert.Equal(t, 9, target.Day())
		assert.Equal(t, 15, target.Hour())
	})

	t.Run("During Saturday (Saturday 10:00 -> Next Monday 15:00)", func(t *testing.T) {
		// 2026-03-07 is Saturday
		now := time.Date(2026, 3, 7, 10, 0, 0, 0, loc)
		target := nextWeekdayTarget(now, 15, 0, 0)

		assert.Equal(t, time.Monday, target.Weekday())
		assert.Equal(t, 9, target.Day())
		assert.Equal(t, 15, target.Hour())
	})

	t.Run("During Sunday (Sunday 10:00 -> Next Monday 15:00)", func(t *testing.T) {
		// 2026-03-08 is Sunday
		now := time.Date(2026, 3, 8, 10, 0, 0, 0, loc)
		target := nextWeekdayTarget(now, 15, 0, 0)

		assert.Equal(t, time.Monday, target.Weekday())
		assert.Equal(t, 9, target.Day())
		assert.Equal(t, 15, target.Hour())
	})
}
