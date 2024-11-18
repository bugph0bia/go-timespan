// timespan: Calculating time spans
package timespan

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// base time
var basetime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)

// Time: Create Time, if you don't want to care about year, month, date, and nanoseconds
func Time(hour int, min int, sec int) time.Time {
	return time.Date(
		basetime.Year(), basetime.Month(), basetime.Day(),
		hour, min, sec,
		basetime.Nanosecond(), time.Local)
}

// Hms: Get Hour, Minite, and Second, if you don't want to care about year, month, date, and nanoseconds
func Hms(t time.Time) (int, int, int) {
	// 基準時刻からの増加分を計算
	d := t.Sub(basetime)
	// 時分秒に分解するための正規化処理
	// 時は24以上を許容する
	norm := func(d time.Duration, base time.Duration) (int, time.Duration) {
		x := d / base
		d -= x * base
		return int(x), d
	}
	hour, d := norm(d, time.Hour)
	min, d := norm(d, time.Minute)
	sec, _ := norm(d, time.Second)
	return hour, min, sec
}

// unit: unit of a time span
type unit struct {
	t [2]time.Time
}

// Start: Return the start time
func (u unit) Start() time.Time {
	if u.t[0].Before(u.t[1]) {
		return u.t[0]
	} else {
		return u.t[1]
	}
}

// End: Return the end time
func (u unit) End() time.Time {
	if u.t[0].After(u.t[1]) {
		return u.t[0]
	} else {
		return u.t[1]
	}
}

// Length: Return the duration
func (u unit) Length() time.Duration {
	return u.End().Sub(u.Start())
}

// IncludesTime: Check if time are included inside
func (u unit) IncludesTime(t time.Time) bool {
	// u.Start <= t <= u.End
	return (u.Start().Before(t) || u.Start().Equal(t)) && (u.End().After(t) || u.End().Equal(t))
}

// Includes: Check if other units are included inside
func (u unit) Includes(other unit) bool {
	// other の始点・終点が u の範囲内
	return u.IncludesTime(other.Start()) && u.IncludesTime(other.End())
}

// Continuous: Check for continuity with other unit
func (u unit) Continuous(other unit) bool {
	// 始点・終点のいずれかがもう一方の範囲内
	return u.IncludesTime(other.Start()) || u.IncludesTime(other.End()) || other.IncludesTime(u.Start()) || other.IncludesTime(u.End())
}

// String: Return string
func (u unit) String() string {
	sh, sm, ss := Hms(u.Start())
	eh, em, es := Hms(u.End())
	st := fmt.Sprintf("%02d:%02d:%02d", sh, sm, ss)
	et := fmt.Sprintf("%02d:%02d:%02d", eh, em, es)
	return fmt.Sprintf("[%s - %s]", st, et)
}

// newUnit: Create new unit
func newUnit(t1 time.Time, t2 time.Time) unit {
	if t1.Before(t2) {
		return unit{t: [2]time.Time{t1, t2}}
	} else {
		return unit{t: [2]time.Time{t2, t1}}
	}
}

// addUnits: Add units
func addUnits(lhs unit, rhs unit) []unit {
	units := make([]unit, 0, 2)
	if lhs.Continuous(rhs) {
		var start time.Time
		if lhs.Start().Before(rhs.Start()) {
			start = lhs.Start()
		} else {
			start = rhs.Start()
		}
		var end time.Time
		if lhs.End().After(rhs.End()) {
			end = lhs.End()
		} else {
			end = rhs.End()
		}
		units = append(units, newUnit(start, end))
	} else {
		units = append(units, lhs, rhs)
	}
	return units
}

// subUnits: Sub units
func subUnits(lhs unit, rhs unit) []unit {
	units := make([]unit, 0, 2)
	switch {
	// 非連続 = 変化なし
	case !lhs.Continuous(rhs):
		units = append(units, lhs)
	// 内包 = 二分割される
	case lhs.Includes(rhs):
		units = append(units, newUnit(lhs.Start(), rhs.Start()), newUnit(rhs.End(), lhs.End()))
	// 逆内包 = 消去
	case rhs.Includes(lhs):
		break
	// 右部分で重なる = 終点が短くなる
	case lhs.Start().Before(rhs.Start()):
		units = append(units, newUnit(lhs.Start(), rhs.Start()))
	// 左部分で重なる = 始点が短くなる
	case lhs.Start().After(rhs.Start()):
		units = append(units, newUnit(rhs.End(), lhs.End()))
	}
	return units
}

// TimeSpan: Holds multiple time span units
type TimeSpan struct {
	units []unit
}

// New: Create a TimeSpan from two times
func New(start time.Time, end time.Time) *TimeSpan {
	var t [2]time.Time
	if start.Before(end) {
		t[0] = start
		t[1] = end
	} else {
		t[0] = end
		t[1] = start
	}
	return &TimeSpan{
		units: []unit{{t: t}},
	}
}

// NewDuration: Create a TimeSpan from a time and a duration
func NewDuration(start time.Time, duration time.Duration) *TimeSpan {
	return New(start, start.Add(duration))
}

// NewBulk: Create a TimeSpan from a slice
func NewBulk(spans [][2]time.Time) *TimeSpan {
	var ts TimeSpan
	for _, span := range spans {
		ts.units = append(ts.units, newUnit(span[0], span[1]))
	}
	ts.normalize()
	return &ts
}

// StartEnd: Return start time and end time
func (ts *TimeSpan) StartEnd() (start time.Time, end time.Time, err error) {
	if ts.Count() == 0 {
		return time.Time{}, time.Time{}, errors.New("having no time span")
	}
	return ts.units[0].Start(), ts.units[ts.Count()-1].End(), nil
}

// Total: Return total duration
func (ts *TimeSpan) Total() time.Duration {
	var d time.Duration
	for _, u := range ts.units {
		d += u.Length()
	}
	return d
}

// Count: Return count of spans
func (ts *TimeSpan) Count() int {
	return len(ts.units)
}

// Raw: Return inner spans in slices
func (ts *TimeSpan) Raw() [][2]time.Time {
	s := [][2]time.Time{}
	for _, u := range ts.units {
		s = append(s, u.t)
	}
	return s
}

// Add: Add TimeSpans
func (ts *TimeSpan) Add(other *TimeSpan) {
	// 末尾に追加して正規化
	ts.units = append(ts.units, other.units...)
	ts.normalize()
}

// Sub: Sub TimeSpans
func (ts *TimeSpan) Sub(other *TimeSpan) {
	// 全組み合わせで減算して正規化
	newUnits := make([]unit, 0, ts.Count()*2)
	for _, u1 := range ts.units {
		for _, u2 := range other.units {
			newUnits = append(newUnits, subUnits(u1, u2)...)
		}
	}
	ts.units = newUnits
	ts.normalize()
}

// IncludesTime: Check if time are included inside
func (ts *TimeSpan) IncludesTime(t time.Time) bool {
	// いずれかの範囲内にあるか
	for _, u := range ts.units {
		if u.IncludesTime(t) {
			return true
		}
	}
	return false
}

// Includes: Check if other TimeSpan are included inside
func (ts *TimeSpan) Includes(other *TimeSpan) bool {
	// other の全ての範囲が ts のいずれかの範囲内にあるか
	for _, u2 := range other.units {
		var check bool
		for _, u1 := range ts.units {
			if u1.Includes(u2) {
				check = true
				break
			}
		}
		if !check {
			return false
		}
	}
	return true
}

// Continuous: Check for continuity with other TimeSpan
func (ts *TimeSpan) Continuous(other *TimeSpan) bool {
	// other のいずれかの範囲が ts のいずれかの範囲と連続しているか
	for _, u2 := range other.units {
		for _, u1 := range ts.units {
			if u1.Continuous(u2) {
				return true
			}
		}
	}
	return false
}

// String: Return string
func (ts *TimeSpan) String() string {
	ss := make([]string, 0, ts.Count())
	for _, u := range ts.units {
		ss = append(ss, u.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Normalize: normalize inner data
func (ts *TimeSpan) normalize() {
	// 始点をキーに昇順ソート
	sort.Slice(ts.units, func(i, j int) bool {
		u1 := &ts.units[i]
		u2 := &ts.units[j]
		return u1.Start().Equal(u2.Start()) || u1.Start().Before(u2.Start())
	})

	// 基準範囲のループ
	for i := 0; i < len(ts.units)-1; i++ {
		// 無効な要素はスキップ
		if ts.units[i].Length() < 0 {
			continue
		}

		// 隣接範囲のループ
		for j := i + 1; j < len(ts.units); j++ {
			// 無効な要素はスキップ
			if ts.units[j].Length() < 0 {
				continue
			}

			// 隣同士の範囲を合成
			us := addUnits(ts.units[i], ts.units[j])
			if len(us) == 1 {
				ts.units[i] = us[0]                             // 合成後の範囲
				ts.units[j] = newUnit(time.Time{}, time.Time{}) // 削除
			} else {
				// 合成されなかった場合は基準を進める
				break
			}
		}
	}

	// 無効な要素を削除
	newUnits := make([]unit, 0, len(ts.units))
	for _, u := range ts.units {
		if u.Length() > 0 {
			newUnits = append(newUnits, u)
		}
	}
	ts.units = newUnits
}
