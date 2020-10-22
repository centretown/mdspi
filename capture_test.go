package main

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func testCapture(t *testing.T) {
	pipeline := make(chan *Measurement)
	monitor := make(chan int)
	go capture(pipeline, monitor)
	m := &Measurement{
		interval:    1,
		accel:       [3]float64{0.0, 0.0, 0.0},
		gyro:        [3]float64{0, 0, 0},
		temperature: 27.0,
	}

	m.interval = .5
	for i := 0; i < 10; i++ {
		pipeline <- m
	}

}

func TestUpdateState(t *testing.T) {
	var dt float64 = .004 // seconds
	// Filter -
	// current state
	X := mat.NewVecDense(4, []float64{
		0.0, // distance x
		0.0, // distance y
		0.0, // velocity x
		0.0, // velocity y
	})
	resetState := func() {
		X = mat.NewVecDense(4, []float64{
			0.0, // distance x
			0.0, // distance y
			0.0, // velocity x
			0.0, // velocity y
		})
	}

	// prediction matrix (n x n)
	A := mat.NewDense(4, 4, []float64{
		1.0, 0.0, dt, 0.0,
		0.0, 1.0, 0.0, dt,
		0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 1.0,
	})
	// control matrix (n x k)
	// change due to acceleration
	B := mat.NewDense(4, 2, []float64{
		0.5 * dt * dt, 0.0,
		0.0, 0.5 * dt * dt,
		dt, 0.0,
		0.0, dt,
	})

	m := &Measurement{
		interval:    1,
		accel:       [3]float64{0.0, 0.0, 0.0},
		gyro:        [3]float64{0, 0, 0},
		temperature: 27.0,
	}

	var stateCycle = func(count float64) {
		fmt.Println("time: ", 0)
		printMatrix([]mat.Matrix{X},
			"      X = ")
		interval := m.interval
		i := float64(0.0)
		for {
			if i >= count {
				break
			}
			i += m.interval
			if i > count {
				m.interval = i - count
			}

			x := updateState(A, B, X, m)
			X.CopyVec(x)

			fmt.Println("time: ", i)
			printMatrix([]mat.Matrix{X},
				"      X = ")
		}
		m.interval = interval
	}

	showMatrices = false
	m.accel[iX] = 1
	m.accel[iY] = .75
	m.interval = 1
	resetState()
	stateCycle(3)
	// resetState()
	// stateCycle(3)
	// stateCycle(3)

	// set velocities to zero, stop
	X.SetVec(2, 0)
	X.SetVec(3, 0)
	// go back
	m.accel[iX] = -m.accel[iX]
	m.accel[iY] = -m.accel[iY]
	m.interval = 1
	stateCycle(3)

	// resetState()
	// m.interval = .4
	// stateCycle()
}

func testAdjust(t *testing.T) {
	dt := .004
	// prediction matrix (n x n)
	A := mat.NewDense(6, 6, []float64{
		1.0, 0.0, 0.0, dt, 0.0, 0.0,
		0.0, 1.0, 0.0, 0.0, dt, 0.0,
		0.0, 0.0, 1.0, 0.0, 0.0, dt,
		0.0, 0.0, 0.0, 1.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 1.0,
	})

	dtt := 0.5 * dt * dt

	B := mat.NewDense(6, 3, []float64{
		dtt, 0.0, 0.0,
		0.0, dtt, 0.0,
		0.0, 0.0, dtt,
		dt, 0.0, 0.0,
		0.0, dt, 0.0,
		0.0, 0.0, dt,
	})
	f := func(dt float64) float64 { return 0.5 * dt * dt }

	printMatrix([]mat.Matrix{A, B})
	dt = .5
	adjustABForInterval(dt, A, B, f)
	printMatrix([]mat.Matrix{A, B})

	// prediction matrix (n x n)
	A = mat.NewDense(4, 4, []float64{
		1.0, 0.0, dt, 0.0,
		0.0, 1.0, 0.0, dt,
		0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 1.0,
	})
	// control matrix (n x k)
	// change due to acceleration
	B = mat.NewDense(4, 2, []float64{
		0.5 * dt * dt, 0.0,
		0.0, 0.5 * dt * dt,
		dt, 0.0,
		0.0, dt,
	})
	printMatrix([]mat.Matrix{A, B})
	dt = 4
	adjustABForInterval(dt, A, B, f)
	printMatrix([]mat.Matrix{A, B})

	X := mat.NewVecDense(6, []float64{
		0.0, // distance x
		0.0, // distance y
		0.0, // distance z
		1.0, // velocity x
		1.0, // velocity y
		1.0, // velocity z
	})
	fmt.Println(X.Dims())
}

func testGetAccel(t *testing.T) {
	m := &Measurement{
		interval:    1,
		accel:       [3]float64{0.1, 0.2, 0.3},
		gyro:        [3]float64{0, 0, 0},
		temperature: 27.0,
	}
	for n := 1; n < 4; n++ {
		u := getAccel(n, m)
		printMatrix([]mat.Matrix{u})
	}
}

func testUpdateState3D(t *testing.T) {
	showMatrices = true
	var dt float64 = .004 // seconds
	// Filter -
	// current state
	X := mat.NewVecDense(6, []float64{
		0.0, // distance x
		0.0, // distance y
		0.0, // distance z
		0.0, // velocity x
		0.0, // velocity y
		0.0, // velocity z
	})
	// prediction matrix (n x n)
	A := mat.NewDense(6, 6, []float64{
		1.0, 0.0, 0.0, dt, 0.0, 0.0,
		0.0, 1.0, 0.0, 0.0, dt, 0.0,
		0.0, 0.0, 1.0, 0.0, 0.0, dt,
		0.0, 0.0, 0.0, 1.0, 0.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0, 0.0, 1.0,
	})
	// control matrix (n x k)
	// change due to acceleration
	dtt := 0.5 * dt * dt
	B := mat.NewDense(6, 3, []float64{
		dtt, 0.0, 0.0,
		0.0, dtt, 0.0,
		0.0, 0.0, dtt,
		dt, 0.0, 0.0,
		0.0, dt, 0.0,
		0.0, 0.0, dt,
	})

	m := &Measurement{
		interval:    1,
		accel:       [3]float64{1.0, 1.0, 1.0},
		gyro:        [3]float64{0, 0, 0},
		temperature: 27.0,
	}

	for i := 0; i < 4; i++ {
		printMatrix([]mat.Matrix{X},
			"X = ")
		x := updateState(A, B, X, m)
		X.CopyVec(x)
		fmt.Printf("%v seconds\n", float64(i+1)*m.interval)
	}
	printMatrix([]mat.Matrix{X},
		"X = ")

}
