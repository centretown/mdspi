package main

import (
	"fmt"
	"strings"
	"time"

	"gonum.org/v1/gonum/mat"
)

var showRaw bool
var showMatrices bool = true

func capture(pipeline <-chan *Measurement, monitor chan int) {
	var (
		estimateError   float64 = 1.0
		meaurementError float64 = .5   // Error in Measurement
		kalmanGain      float64        // Kalman Gain
		estimate        float64 = 30.0 // current estimate

		dt float64 = 0.004 // seconds
		// Filter -
		// current state
		X = mat.NewVecDense(4, []float64{
			0.0, // distance x
			0.0, // distance y
			0.0, // velocity x
			0.0, // velocity y
		})
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
		// measurement matrix (l x n)
		// C = mat.NewDense(4, 1, []float64{0.0, 0.0, 0.0, 0.0})
		// measurement matrix (l x k)
		// D = mat.NewDense(4, 1, []float64{0.0, 0.0, 0.0, 0.0})
	)

	var measurement *Measurement
	show := func(estimate float64) {
		fmt.Printf("dt=%1.6f  ax=%+2.6f ay=%+2.6f az=%+2.6f gx=%+2.6f gy=%+2.6f gz=%+2.6f m=%6.4f e=%6.4f\n",
			// fmt.Printf("dt=%1.6f  ax=%+2.6f ay=%+2.6f az=%+2.6f m=%6.4f e=%6.4f\n",
			measurement.interval,
			measurement.accel[iX],
			measurement.accel[iY],
			measurement.accel[iZ],
			measurement.gyro[iX],
			measurement.gyro[iY],
			measurement.gyro[iZ],
			measurement.temperature, estimate)
	}

	for {
		select {
		case status := <-monitor:
			switch status {
			case MonitorRaw:
				showRaw = !showRaw
			case MonitorMatrices:
				showMatrices = !showMatrices
			}
		case measurement = <-pipeline:
			kalmanGain = estimateError / (estimateError + meaurementError)
			estimate = estimate + kalmanGain*(measurement.temperature-estimate)
			estimateError = (1 - kalmanGain) * estimateError

			x := updateState(A, B, X, measurement)
			X.CopyVec(x)
			if showRaw {
				show(estimate)
			}
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

var spaces = strings.Repeat(" ", 16)
var optPrefix = mat.Prefix(spaces)
var optSqueeze = mat.Squeeze()

func printMatrix(mats []mat.Matrix, labels ...string) {
	labelCount := len(labels)
	label := spaces
	labelWidth := len(spaces)
	for i, m := range mats {
		if i < labelCount {
			label = labels[i]
		} else {
			label = spaces
		}

		labelLength := len(label)
		if labelLength < labelWidth {
			// pad left if short
			label = spaces[:labelWidth-labelLength] + label
		}

		f := mat.Formatted(m, optPrefix, optSqueeze)
		fmt.Printf(label+"%v\n", f)
	}
}

// adjust A and B by where A is r x r and B is r x r/2 matrix
func adjustABForInterval(dt float64, A, B *mat.Dense, f func(dt float64) float64) {
	dtt := f(dt)
	half, _ := A.Dims()
	half /= 2
	for i := 0; i < half; i++ {
		A.Set(i, i+half, dt)
		B.Set(i, i, dtt)
		B.Set(i+half, i, dt)
	}
}

// get acceleration data for n dimemsions
func getAccel(n int, m *Measurement) (u *mat.Dense) {
	u = mat.NewDense(n, 1, nil)
	for i := 0; i < n; i++ {
		u.Set(i, 0, m.accel[i])
	}
	return
}

var acceleration = func(dt float64) float64 { return 0.5 * dt * dt }

// updates state vector
// Xk = A*Xk-1 + B*uk + wk
func updateState(A, B *mat.Dense, X mat.Vector, measurement *Measurement) mat.Vector {
	// adjust time interval
	dt := measurement.interval

	adjustABForInterval(dt, A, B, acceleration)

	var ax, bu, x mat.Dense
	ax.Mul(A, X)

	// u := mat.NewDense(2, 1,
	// 	[]float64{measurement.accel[iX], measurement.accel[iY]})
	// get value for each column of B matrix
	_, n := B.Dims()
	u := getAccel(n, measurement)
	// B . u
	bu.Mul(B, u)

	// Ax + Bu
	x.Add(&ax, &bu)
	// return new state
	// X.CopyVec(x.ColView(0))
	if showMatrices {
		printMatrix([]mat.Matrix{A, X, &ax},
			"A = ",
			"X = ",
			"A * X = ")
		fmt.Println()
		printMatrix([]mat.Matrix{B, u, &bu},
			"B = ",
			"u = ",
			"B * u = ")
		fmt.Println()
		printMatrix([]mat.Matrix{&ax, &bu, &x},
			"ax = ",
			"bu = ",
			"ax + bu = ")
		fmt.Println()
		fmt.Println()
	}
	return x.ColView(0)
}
