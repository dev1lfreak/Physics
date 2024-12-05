package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"math"
	"os"
)

type consts struct {
	objectMass  float64 // масса корабля и пилота
	fuelMass    float64 // масса топлива
	jetSpeed    float64 // скорость выброса частиц
	burnSpeed   float64 // скорость расхода топлива
	g           float64 // ускорение свободного падения
	startSpeed  float64 // начальная скорость
	startHeight float64 // начальная высота
	fallTime    float64 // время падения корабля с выключенным двигателем
	engineTime  float64 // время падения корабля с включенным двигателем
	mass        float64 // общая масса объекта (масса корабля, пилота и топлива)
}

var num consts = consts{
	2150.0, 150.0,
	3660.0, 15.0,
	1.62, 0,
	0, 0,
	0, 2300}

// CountFallSpeed Функция, считающая скорость, которую корабль набирает, пока падает с выключенным двигателем, за t
func CountFallSpeed(t float64) float64 {
	v := num.startSpeed + num.g*t
	return v
}

// CountSpeedWithWorkingEngine Функция, считающая скорость, которую корабль имеет с включенным двигателем в момент t
func CountSpeedWithWorkingEngine(ft, t float64) float64 {
	v := CountFallSpeed(ft) + num.g*t - num.jetSpeed*math.Log((num.mass)/(num.mass-num.burnSpeed*t))
	return v
}

// CountFallHeight Функция, считающая высоту, которую пролетел корабль с выключенным двигателем за время t
func CountFallHeight(t float64) float64 {
	h := num.startSpeed*t + num.g*math.Pow(t, 2)/2.0
	return h
}

// CountHeightWithWorkingEngine Функция, считающая высоту, которую пролетел корабль с включенным двигателем за время t
func CountHeightWithWorkingEngine(ft, t float64) float64 {
	t1 := CountFallSpeed(ft) * t
	t2 := num.g * math.Pow(t, 2) / 2.0
	t3 := math.Log(num.mass) * t
	t4 := (num.mass-num.burnSpeed*t)*math.Log(num.mass-num.burnSpeed*t) -
		num.mass*math.Log(num.mass) + num.burnSpeed*t

	h := t1 + t2 - num.jetSpeed*(t3+(1/num.burnSpeed)*t4)
	return h
}

// CountAccelerationWithWorkingEngine Функция, считающая ускорение корабля с работающим двигателем за время t
func CountAccelerationWithWorkingEngine(t float64) float64 {
	a := num.g - num.jetSpeed*(num.burnSpeed/(num.mass-num.burnSpeed*t))
	return a
}

// FindVariableTime Функция, которая решает систему уравнений путем перебора значений
func FindVariableTime() {
	for i := 1.0; CountFallHeight(i) <= num.startHeight; i += 0.001 {
		for j := 0.5; j <= 10; j += 0.05 {
			// Проверяем, являются ли данные значения искомыми, сравнивая суммарную высоту с начальной и
			//проверяя вхождение конечной скорости в отрезок [0;3]
			if (math.Abs(CountHeightWithWorkingEngine(i, j)+CountFallHeight(i)-num.startHeight) <= 0.5) &&
				(0.0 <= CountSpeedWithWorkingEngine(i, j) && CountSpeedWithWorkingEngine(i, j) <= 3.0) {
				num.fallTime = i
				num.engineTime = j
				break
			}
		}
	}
}

func generateSpeedData(f func(ft float64) float64, g func(ft, t float64) float64, start, end, step float64) ([]opts.LineData, []float64) {
	var xVals []float64
	var yVals []opts.LineData

	for x := start; x <= num.fallTime; x += step {
		y := f(x)
		xVals = append(xVals, x)
		yVals = append(yVals, opts.LineData{Value: y})
	}

	for x := start; x <= num.engineTime; x += step {
		y := g(num.fallTime, x)
		xVals = append(xVals, num.fallTime+x)
		yVals = append(yVals, opts.LineData{Value: y})
	}
	return yVals, xVals
}

func generateHeightData(f func(ft float64) float64, g func(ft, t float64) float64, start, end, step float64) ([]opts.LineData, []float64) {
	var xVals []float64
	var yVals []opts.LineData

	for x := start; x <= num.fallTime; x += step {
		y := f(x)
		xVals = append(xVals, x)
		yVals = append(yVals, opts.LineData{Value: y})
	}

	for x := start; x <= num.engineTime; x += step {
		y := CountFallHeight(num.fallTime) + g(num.fallTime, x)
		xVals = append(xVals, num.fallTime+x)
		yVals = append(yVals, opts.LineData{Value: y})
	}
	return yVals, xVals
}

func generateAccelerationData(f func(t float64) float64, start, end, step float64) ([]opts.LineData, []float64) {
	var xVals []float64
	var yVals []opts.LineData

	for x := start; x <= num.fallTime; x += step {
		y := num.g
		xVals = append(xVals, x)
		yVals = append(yVals, opts.LineData{Value: y})
	}

	for x := start; x <= num.engineTime; x += step {
		y := f(x)
		xVals = append(xVals, num.fallTime+x)
		yVals = append(yVals, opts.LineData{Value: y})
	}
	return yVals, xVals
}

func main() {
	var startS, startH float64
	fmt.Scanln(&startS, &startH)
	num.startSpeed = startS
	num.startHeight = startH
	FindVariableTime()
	fmt.Println(CountSpeedWithWorkingEngine(num.fallTime, num.engineTime))

	start := 0.0
	end := num.engineTime + num.fallTime
	step := 0.01

	// Generate data for the chart.
	yVals, xVals := generateSpeedData(CountFallSpeed, CountSpeedWithWorkingEngine, start, end, step)

	// Create a line chart.
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "График изменения скорости"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "t, с"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "V, м/c"}),
	)

	// Set X and Y axis data.
	line.SetXAxis(xVals).
		AddSeries("V(t)", yVals).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{}))

	// Save the chart to an HTML file.
	f, _ := os.Create("graph_speed_change.html")
	defer f.Close()
	line.Render(f)

	yVals, xVals = generateHeightData(CountFallHeight, CountHeightWithWorkingEngine, start, end, step)

	// Create a line chart.
	line = charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "График высоты"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "t, с"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "H, м"}),
	)

	// Set X and Y axis data.
	line.SetXAxis(xVals).
		AddSeries("H(t)", yVals).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{}))

	// Save the chart to an HTML file.
	f, _ = os.Create("graph_height.html")
	defer f.Close()
	line.Render(f)

	yVals, xVals = generateAccelerationData(CountAccelerationWithWorkingEngine, start, end, step)

	// Create a line chart.
	line = charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "График изменения ускорения"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "t, с"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "a, м/c²"}),
	)

	// Set X and Y axis data.
	line.SetXAxis(xVals).
		AddSeries("a(t)", yVals).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{}))

	// Save the chart to an HTML file.
	f, _ = os.Create("graph_acceleration.html")
	defer f.Close()
	line.Render(f)
}
