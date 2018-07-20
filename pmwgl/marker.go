package pmwgl

import "math"

const (
	unitDetail = 20
	unitRadius = 20
	height     = -100
)

var unitCircle = []float32{
	0, 0, 0,
}

var unitCirclePoints int

func init() {
	for i := 0; i < unitDetail; i++ {
		unitCircle = append(unitCircle, float32(unitRadius*math.Cos((float64(i)/unitDetail)*2.0*math.Pi)))
		unitCircle = append(unitCircle, float32(unitRadius*math.Sin((float64(i)/unitDetail)*2.0*math.Pi)))
		unitCircle = append(unitCircle, height)
	}
	unitCircle = append(unitCircle, unitCircle[3], unitCircle[4], height)
	unitCirclePoints = len(unitCircle) / 3
}
