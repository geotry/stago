package compute

import (
	"time"
)

const PI = 3.14159265

func Clamp(v float64, min float64, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Return a value between {min} and {max}, scaled at {scale}, at point {t}
func LinearStep(min float64, max float64, scale float64, point float64) float64 {
	return min + (point * (1 / scale) * (max - min))
}

// {min} {max} the range of the value
// {value} the initial value
// {duration} the duration of the animation
// {time} the key frame
func LinerarLoop(min float64, max float64, value float64, duration time.Duration, time time.Duration) float64 {
	r := (max - min)
	d := int64(time) % int64(duration)
	s := (float64(d)) / float64(duration)
	nv := value + (r * s)

	if nv > max {
		return (max - nv)
	}
	return nv
}
