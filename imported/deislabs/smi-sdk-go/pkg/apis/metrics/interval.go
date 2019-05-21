package metrics

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultWindow = 30 * time.Second

// Interval defines the time interval from which metrics were collected
type Interval struct {
	Timestamp metav1.Time     `json:"timestamp"`
	Window    metav1.Duration `json:"window"`
}

func NewInterval() *Interval {
	return &Interval{
		Timestamp: metav1.NewTime(time.Now()),
		Window:    metav1.Duration{Duration: defaultWindow},
	}
}

// String returns a formatted string representation of this struct
func (i *Interval) String() string {
	return fmt.Sprintf("%#v", i)
}
