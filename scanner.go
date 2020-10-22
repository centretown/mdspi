package main

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

// ScanLines from a reader and output to each line
func ScanLines(rw io.ReadWriter, out chan<- string, in <-chan string) (err error) {
	bufReader := bufio.NewReaderSize(rw, 512)
	bufWriter := bufio.NewWriterSize(rw, 1)

	var (
		str string
	)

	for {
		select {
		case sendChar := <-in:
			bufWriter.WriteByte(sendChar[0])
		default:
			str, err = bufReader.ReadString('\n')
			if err != nil {
				return
			}
			out <- str
		}
		time.Sleep(time.Millisecond)
	}
}

const (
	iX int = iota
	iY
	iZ
)

// GYROSCOPE SENSITIVITY
// Full-Scale Range      Sensitivity Scale Factor
// FS_SEL=0 ±250 º/s     FS_SEL=0 131 LSB/(º/s)
// FS_SEL=1 ±500 º/s     FS_SEL=1 65.5 LSB/(º/s)
// FS_SEL=2 ±1000 º/s    FS_SEL=2 32.8 LSB/(º/s)
// FS_SEL=3 ±2000 º/s    FS_SEL=3 16.4 LSB/(º/s)
// Gyroscope ADC Word Length 16 bits
//
// Sensitivity Scale Factor Tolerance 25°C min:-3% max:+3%
// Sensitivity Scale Factor Variation Over
// Temperature
// ±2 %
// Nonlinearity Best fit straight line; 25°C 0.2 %
// Cross-Axis Sensitivity ±2 %

const (
	// G - acceleration due to gravity
	G = -9.81
	// GyroSensitivity gyro Sensitivity Scale Factor
	// FS_SEL=0 ±250 º/s     FS_SEL=0 131 LSB/(º/s)
	GyroSensitivity float64 = 131.1
	// AcclSensitivity accelerometer Sensitivity Scale Factor
	// FS_SEL=0±2g FS_SEL=016,384 LSB/g
	AcclSensitivity float64 = 16384.0 // meters/second**2
	// OverFlowCorrection :
	// micros value read from a 32 bit value which
	// will overflow after approximately 70 minutes
	OverFlowCorrection = 0x100000000
	// MicrosPerSecond conversion factor
	MicrosPerSecond float64 = 1000000.0
)

// Measurement -
type Measurement struct {
	lasttime    float64    // secs
	interval    float64    // secs
	accel       [3]float64 // meters per second ** 2
	gyro        [3]float64 // degrees per second
	temperature float64
}

// Measure -
func Measure(in <-chan string, pipeline chan<- *Measurement) {
	var currentMicros, previousMicros uint64
	var counter int = 200

	for {
		select {
		case msg := <-in:
			m := &Measurement{}
			fmt.Sscanln(msg,
				&currentMicros,
				&m.accel[iX], //ax
				&m.accel[iY], //ay
				&m.accel[iZ], //az
				&m.gyro[iX],  //gx
				&m.gyro[iY],  //gy
				&m.gyro[iZ],  //gZ
				&m.temperature)
			// throw the first few away
			if counter > 0 {
				counter--
				previousMicros = currentMicros
				continue
			}

			if currentMicros > previousMicros {
				m.interval = float64(currentMicros-previousMicros) / MicrosPerSecond
			} else {
				// correct for overflow
				m.interval = float64(OverFlowCorrection+currentMicros-previousMicros) / MicrosPerSecond
			}
			previousMicros = currentMicros

			// convert g's to meters/second**2
			m.accel[iX] = (m.accel[iX] * G) / AcclSensitivity
			m.accel[iY] = (m.accel[iY] * G) / AcclSensitivity
			m.accel[iZ] = (m.accel[iZ] * G) / AcclSensitivity

			m.gyro[iX] /= GyroSensitivity
			m.gyro[iY] /= GyroSensitivity
			m.gyro[iZ] /= GyroSensitivity
			// Temperature in degrees C = (TEMP_OUT Register Value as a signed quantity)/340 + 36.53
			m.temperature = float64(m.temperature)/340.0 + 36.53

			pipeline <- m
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
