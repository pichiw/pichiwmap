package pmwgl

import "math"

const (
	unitDetail = 20
	unitRadius = 1
)

var unitCircle = []float32{
	0, 0, -1,
}

var unitCirclePoints int

func init() {
	for i := 0; i < unitDetail; i++ {
		unitCircle = append(unitCircle, float32(unitRadius*math.Cos((float64(i)/unitDetail)*2.0*math.Pi)))
		unitCircle = append(unitCircle, float32(unitRadius*math.Sin((float64(i)/unitDetail)*2.0*math.Pi)))
		unitCircle = append(unitCircle, 0)
	}
	unitCircle = append(unitCircle, unitCircle[3], unitCircle[4], 0)
	unitCirclePoints = len(unitCircle) / 3
}
