package filters

import "time"

//go:generate go run ./cmd file.go

//genQuery
type Point struct {
	X, Y int
}

//genQuery
type TestFn struct {
	price  int
	Name   string
	Point  *Point
	expire *time.Time
}

func TestfnGlobal(a, b int) int {
	return a + b
}

func (t *TestFn) Price() int {
	print("fn")
	return 23
}

func (t *TestFn) Expire() time.Time {
	return time.Now()
}

func (t *TestFn) Somemethod(pt Point) {
	print(t.price, ' ', t.Name, ' ', t.Point)
}

func (t TestFn) Somemetho3d(pt Point) {
	print(t.price, ' ', t.Name, ' ', t.Point)
}
