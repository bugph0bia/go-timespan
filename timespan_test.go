package timespan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	start := Time(1, 20, 30)
	end := Time(2, 30, 40)

	ts := New(start, end)
	s, e, err := ts.StartEnd()

	assert.Equal(t, start, s)
	assert.Equal(t, end, e)
	assert.NoError(t, err)
}

func Test_NewDuration(t *testing.T) {
	start := Time(1, 0, 0)
	duration := 70 * time.Minute

	ts := NewDuration(start, duration)
	s, e, err := ts.StartEnd()

	assert.Equal(t, start, s)
	assert.Equal(t, start.Add(duration), e)
	assert.NoError(t, err)
}

func Test_NewBulk(t *testing.T) {
	bulk := [][2]time.Time{
		{Time(3, 10, 0), Time(4, 20, 0)},
		{Time(1, 30, 0), Time(2, 40, 0)},
	}

	ts := NewBulk(bulk)
	s, e, err := ts.StartEnd()

	assert.Equal(t, Time(1, 30, 0), s)
	assert.Equal(t, Time(4, 20, 0), e)
	assert.NoError(t, err)
}

func Test_Start(t *testing.T) {
	// 上でテスト済み
	t.Skip()
}

func Test_End(t *testing.T) {
	// 上でテスト済み
	t.Skip()
}

func Test_Total(t *testing.T) {
	start1 := Time(1, 0, 0)
	end1 := Time(2, 0, 0)
	start2 := Time(1, 30, 0) // 上の区間と重複あり
	end2 := Time(2, 30, 0)
	start3 := Time(4, 0, 0)
	end3 := Time(5, 0, 0)

	ts := New(start1, end1)
	ts.Add(New(start2, end2))
	ts.Add(New(start3, end3))

	d := end2.Sub(start1) + end3.Sub(start3)
	assert.Equal(t, d, ts.Total())
}

func Test_Count(t *testing.T) {
	start1 := Time(1, 0, 0)
	end1 := Time(2, 0, 0)
	start2 := Time(1, 30, 0) // 上の区間と重複あり
	end2 := Time(2, 30, 0)
	start3 := Time(4, 0, 0)
	end3 := Time(5, 0, 0)

	ts := New(start1, end1)
	ts.Add(New(start2, end2))
	ts.Add(New(start3, end3))

	assert.Equal(t, 2, ts.Count())
}

func Test_Slice(t *testing.T) {
	// 下でテスト済み
	t.Skip()
}

func Test_Add(t *testing.T) {
	// 上でテスト済み
	t.Skip()
}

func Test_Sub(t *testing.T) {
	start1 := Time(1, 0, 0)
	end1 := Time(2, 0, 0)
	start2 := Time(1, 30, 0) // 上の区間と重複あり
	end2 := Time(1, 40, 0)

	ts := New(start1, end1)
	ts.Sub(New(start2, end2))

	s := [][2]time.Time{
		{
			start1,
			start2,
		},
		{
			end2,
			end1,
		},
	}
	assert.Equal(t, s, ts.Raw())
}

func Test_IsContinuousAndInclude(t *testing.T) {
	ts1 := New(Time(1, 0, 0), Time(2, 0, 0))
	ts1.Add(New(Time(3, 0, 0), Time(4, 0, 0)))

	// 一致
	ts2 := New(Time(1, 0, 0), Time(2, 0, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, true, ts1.Includes(ts2))

	// 内包
	ts2 = New(Time(1, 10, 0), Time(1, 20, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, true, ts1.Includes(ts2))
	ts2 = New(Time(1, 10, 0), Time(1, 10, 0)) // 長さ0
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, true, ts1.Includes(ts2))
	assert.Equal(t, true, ts1.IncludesTime(Time(1, 10, 0)))

	// 内接
	ts2 = New(Time(1, 0, 0), Time(1, 20, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, true, ts1.Includes(ts2))
	ts2 = New(Time(1, 0, 0), Time(1, 0, 0)) // 長さ0
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, true, ts1.Includes(ts2))
	assert.Equal(t, true, ts1.IncludesTime(Time(1, 0, 0)))

	// 重複
	ts2 = New(Time(1, 50, 0), Time(2, 20, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))
	ts2 = New(Time(1, 50, 0), Time(3, 20, 0)) // 2つの区間にまたがる
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))

	// 連続
	ts2 = New(Time(2, 0, 0), Time(2, 30, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))
	ts2 = New(Time(2, 0, 0), Time(3, 0, 0)) // 2つの区間のちょうど間を埋める
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))

	// 外包
	ts2 = New(Time(0, 0, 0), Time(5, 0, 0))
	assert.Equal(t, true, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))

	// 独立
	ts2 = New(Time(4, 30, 0), Time(4, 40, 0))
	assert.Equal(t, false, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))
	ts2 = New(Time(2, 20, 0), Time(2, 30, 0)) // 2つの区間の間
	assert.Equal(t, false, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))
	ts2 = New(Time(5, 0, 0), Time(5, 0, 0)) // 長さ0
	assert.Equal(t, false, ts1.Continuous(ts2))
	assert.Equal(t, false, ts1.Includes(ts2))
	assert.Equal(t, false, ts1.IncludesTime(Time(5, 0, 0)))
}

func Test_String(t *testing.T) {
	ts := New(Time(1, 10, 10), Time(2, 20, 20))
	ts.Add(New(Time(15, 10, 30), Time(35, 20, 40)))
	assert.Equal(t, "[[01:10:10 - 02:20:20], [15:10:30 - 35:20:40]]", ts.String())
}

func Test_Zero(t *testing.T) {
	var ts TimeSpan
	assert.Equal(t, time.Duration(0), ts.Total())
	assert.Equal(t, 0, ts.Count())
}

func Test_Complex(t *testing.T) {
	ts1 := New(Time(1, 0, 0), Time(100, 0, 0))
	ts2 := New(Time(3, 30, 0), Time(10, 0, 0))
	ts3 := New(Time(30, 0, 0), Time(40, 10, 0))
	ts4 := New(Time(0, 20, 0), Time(20, 0, 0))

	ts1.Sub(ts2)
	ts1.Sub(ts3)
	ts1.Add(ts4)

	want := [][2]time.Time{
		{Time(0, 20, 0), Time(30, 0, 0)},
		{Time(40, 10, 0), Time(100, 0, 0)},
	}
	got := ts1.Raw()
	assert.Equal(t, want, got)
}
