package compute

import (
	"math"
	"testing"
)

func TestScale(t *testing.T) {
	m := NewMatrix4()

	m.Scale(Size{X: 2, Y: 0, Z: 2})

	expected := []float64{
		2, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 2, 0,
		0, 0, 0, 1,
	}

	if len(m.Out) != len(expected) {
		t.Errorf("expected matrix size to be %v, got %v", len(expected), len(m.Out))
		return
	}

	for i := range m.Out {
		if m.Out[i] != expected[i] {
			t.Errorf("expected [%v] to be %v, got %v", i, expected[i], m.Out[i])
		}
	}
}

func TestTranslate(t *testing.T) {
	m := NewMatrix4()

	m.Translate(Point{X: 2, Y: 0, Z: 2})

	expected := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		2, 0, 2, 1,
	}

	if len(m.Out) != len(expected) {
		t.Errorf("expected matrix size to be %v, got %v", len(expected), len(m.Out))
		return
	}

	for i := range m.Out {
		if m.Out[i] != expected[i] {
			t.Errorf("expected [%v] to be %v, got %v", i, expected[i], m.Out[i])
		}
	}
}

func TestRotate(t *testing.T) {
	m := NewMatrix4()

	m.Rotate(Rotation{X: math.Pi / 2, Y: math.Pi / 2, Z: math.Pi / 2})

	if len(m.Out) != 16 {
		t.Errorf("expected matrix size to be %v, got %v", 16, len(m.Out))
	}
}

func TestScaleAndTranslate(t *testing.T) {
	m := NewMatrix4()

	m.Scale(Size{X: 1.2, Y: 1, Z: 3}).Translate(Point{X: -2, Y: 1, Z: 0})

	expected := []float64{
		1.2, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 3, 0,
		-2, 1, 0, 1,
	}

	if len(m.Out) != len(expected) {
		t.Errorf("expected matrix size to be %v, got %v", len(expected), len(m.Out))
		return
	}

	for i := range m.Out {
		if m.Out[i] != expected[i] {
			t.Errorf("expected [%v] to be %v, got %v", i, expected[i], m.Out[i])
		}
	}
}

func TestEquals(t *testing.T) {
	m := NewMatrix4()

	equals := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	if m.Equals(equals) != true {
		t.Errorf("expected matrix to be equal")
	}

	notEquals := []float64{
		0, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	if m.Equals(notEquals) != false {
		t.Errorf("expected matrix to not be equal")
	}

	notSameSize := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
	}

	if m.Equals(notSameSize) != false {
		t.Errorf("expected matrix to not be equal")
	}
}

func TestLookAt(t *testing.T) {
	m := NewMatrix4()

	m.LookAt(Point{X: 0, Y: 0, Z: 0}, Point{X: 0, Y: 0, Z: 1})

	expected := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, -1, 1,
	}

	if m.Equals(expected) != true {
		t.Errorf("expected matrix to be equal: %v", m)
	}

	m.Reset()
	m.LookAt(
		Point{X: 0, Y: 0, Z: 2},
		Point{X: 0, Y: 0, Z: 3},
	)

	expected = []float64{
		-1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, -1, 0,
		0, 0, 0, 1,
	}

	if m.Equals(expected) != true {
		t.Errorf("expected matrix to be equal: %v", m)
	}
}
