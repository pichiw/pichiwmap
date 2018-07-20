package pmwgl

import (
	"math"
)

// Coord represents a coordinate
type Coord struct {
	X float32
	Y float32
	Z float32
	W float32
}

// Normalize normalizes a coord as a vector
func (c Coord) Normalize() Coord {
	length := float32(math.Sqrt(float64((c.X * c.X) + (c.Y * c.Y) + (c.Z * c.Z))))
	if length < 0.00001 {
		return c
	}

	return Coord{
		X: c.X / length,
		Y: c.Y / length,
		Z: c.Z / length,
	}
}

// Sub subtracts c from b
func (c Coord) Sub(b Coord) Coord {
	return Coord{
		X: c.X - b.X,
		Y: c.Y - b.Y,
		Z: c.Z - b.Z,
	}
}

// Cross product of c and b as vectors
func (c Coord) Cross(b Coord) Coord {
	return Coord{
		X: c.Y*b.Z - c.Z*b.Y,
		Y: c.Z*b.X - c.X*b.Z,
		Z: c.X*b.Y - c.Y*b.X,
	}
}

const (
	mi0x0 = iota
	mi0x1
	mi0x2
	mi0x3
	mi1x0
	mi1x1
	mi1x2
	mi1x3
	mi2x0
	mi2x1
	mi2x2
	mi2x3
	mi4x0
	mi4x1
	mi4x2
	mi4x3
)

const (
	ma = iota
	mb
	mc
	md
	me
	mf
	mg
	mh
	mi
	mj
	mk
	ml
	mn
	mo
	mp
)

// Matrix4 represents a 4x4 matrix
type Matrix4 [16]float32

// MultiplyCoord will scale a coordinate
func (m Matrix4) MultiplyCoord(c Coord) Coord {
	return Coord{
		X: (m[ma] * c.X) + (m[mb] * c.Y) + (m[mc] * c.Z) + (m[md] * c.W),
		Y: (m[me] * c.X) + (m[mf] * c.Y) + (m[mg] * c.Z) + (m[mh] * c.W),
		Z: (m[mi] * c.X) + (m[mj] * c.Y) + (m[mk] * c.Z) + (m[ml] * c.W),
		W: (m[ma] * c.X) + (m[mn] * c.Y) + (m[mo] * c.Z) + (m[mp] * c.W),
	}
}

// TransformVector transforms the coord by the matrix
func (m Matrix4) TransformVector(c Coord) Coord {
	v := []float32{c.X, c.Y, c.Z, c.W}
	dst := []float32{0, 0, 0, 0}

	for i := 0; i < 4; i++ {
		dst[i] = 0
		for j := 0; j < 4; j++ {
			dst[i] += (v[j] * m[j*4+i])
		}
	}
	return Coord{
		X: dst[0],
		Y: dst[1],
		Z: dst[2],
		W: dst[3],
	}
}

// Multiply multiplies the receiver by matrix "by" and returns the result
func (m Matrix4) Multiply(by Matrix4) Matrix4 {
	var a00 = m[0*4+0]
	var a01 = m[0*4+1]
	var a02 = m[0*4+2]
	var a03 = m[0*4+3]
	var a10 = m[1*4+0]
	var a11 = m[1*4+1]
	var a12 = m[1*4+2]
	var a13 = m[1*4+3]
	var a20 = m[2*4+0]
	var a21 = m[2*4+1]
	var a22 = m[2*4+2]
	var a23 = m[2*4+3]
	var a30 = m[3*4+0]
	var a31 = m[3*4+1]
	var a32 = m[3*4+2]
	var a33 = m[3*4+3]
	var b00 = by[0*4+0]
	var b01 = by[0*4+1]
	var b02 = by[0*4+2]
	var b03 = by[0*4+3]
	var b10 = by[1*4+0]
	var b11 = by[1*4+1]
	var b12 = by[1*4+2]
	var b13 = by[1*4+3]
	var b20 = by[2*4+0]
	var b21 = by[2*4+1]
	var b22 = by[2*4+2]
	var b23 = by[2*4+3]
	var b30 = by[3*4+0]
	var b31 = by[3*4+1]
	var b32 = by[3*4+2]
	var b33 = by[3*4+3]
	return Matrix4{
		b00*a00 + b01*a10 + b02*a20 + b03*a30,
		b00*a01 + b01*a11 + b02*a21 + b03*a31,
		b00*a02 + b01*a12 + b02*a22 + b03*a32,
		b00*a03 + b01*a13 + b02*a23 + b03*a33,
		b10*a00 + b11*a10 + b12*a20 + b13*a30,
		b10*a01 + b11*a11 + b12*a21 + b13*a31,
		b10*a02 + b11*a12 + b12*a22 + b13*a32,
		b10*a03 + b11*a13 + b12*a23 + b13*a33,
		b20*a00 + b21*a10 + b22*a20 + b23*a30,
		b20*a01 + b21*a11 + b22*a21 + b23*a31,
		b20*a02 + b21*a12 + b22*a22 + b23*a32,
		b20*a03 + b21*a13 + b22*a23 + b23*a33,
		b30*a00 + b31*a10 + b32*a20 + b33*a30,
		b30*a01 + b31*a11 + b32*a21 + b33*a31,
		b30*a02 + b31*a12 + b32*a22 + b33*a32,
		b30*a03 + b31*a13 + b32*a23 + b33*a33,
	}
}

// XRotate rotates the matrix around the x axis
func (m Matrix4) XRotate(angle float64) Matrix4 {
	var dst Matrix4
	var m10 = m[4]
	var m11 = m[5]
	var m12 = m[6]
	var m13 = m[7]
	var m20 = m[8]
	var m21 = m[9]
	var m22 = m[10]
	var m23 = m[11]
	var c = float32(math.Cos(angle))
	var s = float32(math.Sin(angle))

	dst[0] = m[0]
	dst[1] = m[1]
	dst[2] = m[2]
	dst[3] = m[3]
	dst[4] = c*m10 + s*m20
	dst[5] = c*m11 + s*m21
	dst[6] = c*m12 + s*m22
	dst[7] = c*m13 + s*m23
	dst[8] = c*m20 - s*m10
	dst[9] = c*m21 - s*m11
	dst[10] = c*m22 - s*m12
	dst[11] = c*m23 - s*m13
	dst[12] = m[12]
	dst[13] = m[13]
	dst[14] = m[14]
	dst[15] = m[15]

	return dst
}

// YRotate rotates the matrix around the y axis
func (m Matrix4) YRotate(angle float64) Matrix4 {
	var dst Matrix4

	var m00 = m[0*4+0]
	var m01 = m[0*4+1]
	var m02 = m[0*4+2]
	var m03 = m[0*4+3]
	var m20 = m[2*4+0]
	var m21 = m[2*4+1]
	var m22 = m[2*4+2]
	var m23 = m[2*4+3]
	var c = float32(math.Cos(angle))
	var s = float32(math.Sin(angle))

	dst[0] = c*m00 - s*m20
	dst[1] = c*m01 - s*m21
	dst[2] = c*m02 - s*m22
	dst[3] = c*m03 - s*m23
	dst[4] = m[4]
	dst[5] = m[5]
	dst[6] = m[6]
	dst[7] = m[7]
	dst[8] = c*m20 + s*m00
	dst[9] = c*m21 + s*m01
	dst[10] = c*m22 + s*m02
	dst[11] = c*m23 + s*m03
	dst[12] = m[12]
	dst[13] = m[13]
	dst[14] = m[14]
	dst[15] = m[15]

	return dst
}

// ZRotate rotates the matrix around the z axis
func (m Matrix4) ZRotate(angle float64) Matrix4 {
	var dst Matrix4

	var m00 = m[0*4+0]
	var m01 = m[0*4+1]
	var m02 = m[0*4+2]
	var m03 = m[0*4+3]
	var m10 = m[1*4+0]
	var m11 = m[1*4+1]
	var m12 = m[1*4+2]
	var m13 = m[1*4+3]
	var c = float32(math.Cos(angle))
	var s = float32(math.Sin(angle))

	dst[0] = c*m00 + s*m10
	dst[1] = c*m01 + s*m11
	dst[2] = c*m02 + s*m12
	dst[3] = c*m03 + s*m13
	dst[4] = c*m10 - s*m00
	dst[5] = c*m11 - s*m01
	dst[6] = c*m12 - s*m02
	dst[7] = c*m13 - s*m03
	dst[8] = m[8]
	dst[9] = m[9]
	dst[10] = m[10]
	dst[11] = m[11]
	dst[12] = m[12]
	dst[13] = m[13]
	dst[14] = m[14]
	dst[15] = m[15]

	return dst
}

// Scale scales the receiver by x, y, and z
func (m Matrix4) Scale(x, y, z float32) Matrix4 {
	var dst Matrix4
	dst[0] = x * m[0*4+0]
	dst[1] = x * m[0*4+1]
	dst[2] = x * m[0*4+2]
	dst[3] = x * m[0*4+3]
	dst[4] = y * m[1*4+0]
	dst[5] = y * m[1*4+1]
	dst[6] = y * m[1*4+2]
	dst[7] = y * m[1*4+3]
	dst[8] = z * m[2*4+0]
	dst[9] = z * m[2*4+1]
	dst[10] = z * m[2*4+2]
	dst[11] = z * m[2*4+3]
	dst[12] = m[12]
	dst[13] = m[13]
	dst[14] = m[14]
	dst[15] = m[15]
	return dst
}

// ZFactor scales the receiver by x, y, and z
func (m Matrix4) ZFactor(z float32) Matrix4 {
	var dst Matrix4
	copy(dst[:], m[:])
	m[11] = z
	return dst
}

// Translate translates the receiver by x, y, and z
func (m Matrix4) Translate(x, y, z float32) Matrix4 {
	var dst Matrix4

	var m00 = m[0]
	var m01 = m[1]
	var m02 = m[2]
	var m03 = m[3]
	var m10 = m[1*4+0]
	var m11 = m[1*4+1]
	var m12 = m[1*4+2]
	var m13 = m[1*4+3]
	var m20 = m[2*4+0]
	var m21 = m[2*4+1]
	var m22 = m[2*4+2]
	var m23 = m[2*4+3]
	var m30 = m[3*4+0]
	var m31 = m[3*4+1]
	var m32 = m[3*4+2]
	var m33 = m[3*4+3]

	dst[0] = m00
	dst[1] = m01
	dst[2] = m02
	dst[3] = m03
	dst[4] = m10
	dst[5] = m11
	dst[6] = m12
	dst[7] = m13
	dst[8] = m20
	dst[9] = m21
	dst[10] = m22
	dst[11] = m23
	dst[12] = m00*x + m10*y + m20*z + m30
	dst[13] = m01*x + m11*y + m21*z + m31
	dst[14] = m02*x + m12*y + m22*z + m32
	dst[15] = m03*x + m13*y + m23*z + m33

	return dst
}

// Inverse calculates the inverse of the matrix
func (m Matrix4) Inverse() Matrix4 {
	var dst Matrix4
	var m00 = m[0*4+0]
	var m01 = m[0*4+1]
	var m02 = m[0*4+2]
	var m03 = m[0*4+3]
	var m10 = m[1*4+0]
	var m11 = m[1*4+1]
	var m12 = m[1*4+2]
	var m13 = m[1*4+3]
	var m20 = m[2*4+0]
	var m21 = m[2*4+1]
	var m22 = m[2*4+2]
	var m23 = m[2*4+3]
	var m30 = m[3*4+0]
	var m31 = m[3*4+1]
	var m32 = m[3*4+2]
	var m33 = m[3*4+3]
	var tmp_0 = m22 * m33
	var tmp_1 = m32 * m23
	var tmp_2 = m12 * m33
	var tmp_3 = m32 * m13
	var tmp_4 = m12 * m23
	var tmp_5 = m22 * m13
	var tmp_6 = m02 * m33
	var tmp_7 = m32 * m03
	var tmp_8 = m02 * m23
	var tmp_9 = m22 * m03
	var tmp_10 = m02 * m13
	var tmp_11 = m12 * m03
	var tmp_12 = m20 * m31
	var tmp_13 = m30 * m21
	var tmp_14 = m10 * m31
	var tmp_15 = m30 * m11
	var tmp_16 = m10 * m21
	var tmp_17 = m20 * m11
	var tmp_18 = m00 * m31
	var tmp_19 = m30 * m01
	var tmp_20 = m00 * m21
	var tmp_21 = m20 * m01
	var tmp_22 = m00 * m11
	var tmp_23 = m10 * m01

	var t0 = (tmp_0*m11 + tmp_3*m21 + tmp_4*m31) -
		(tmp_1*m11 + tmp_2*m21 + tmp_5*m31)
	var t1 = (tmp_1*m01 + tmp_6*m21 + tmp_9*m31) -
		(tmp_0*m01 + tmp_7*m21 + tmp_8*m31)
	var t2 = (tmp_2*m01 + tmp_7*m11 + tmp_10*m31) -
		(tmp_3*m01 + tmp_6*m11 + tmp_11*m31)
	var t3 = (tmp_5*m01 + tmp_8*m11 + tmp_11*m21) -
		(tmp_4*m01 + tmp_9*m11 + tmp_10*m21)

	var d = 1.0 / (m00*t0 + m10*t1 + m20*t2 + m30*t3)

	dst[0] = d * t0
	dst[1] = d * t1
	dst[2] = d * t2
	dst[3] = d * t3
	dst[4] = d * ((tmp_1*m10 + tmp_2*m20 + tmp_5*m30) -
		(tmp_0*m10 + tmp_3*m20 + tmp_4*m30))
	dst[5] = d * ((tmp_0*m00 + tmp_7*m20 + tmp_8*m30) -
		(tmp_1*m00 + tmp_6*m20 + tmp_9*m30))
	dst[6] = d * ((tmp_3*m00 + tmp_6*m10 + tmp_11*m30) -
		(tmp_2*m00 + tmp_7*m10 + tmp_10*m30))
	dst[7] = d * ((tmp_4*m00 + tmp_9*m10 + tmp_10*m20) -
		(tmp_5*m00 + tmp_8*m10 + tmp_11*m20))
	dst[8] = d * ((tmp_12*m13 + tmp_15*m23 + tmp_16*m33) -
		(tmp_13*m13 + tmp_14*m23 + tmp_17*m33))
	dst[9] = d * ((tmp_13*m03 + tmp_18*m23 + tmp_21*m33) -
		(tmp_12*m03 + tmp_19*m23 + tmp_20*m33))
	dst[10] = d * ((tmp_14*m03 + tmp_19*m13 + tmp_22*m33) -
		(tmp_15*m03 + tmp_18*m13 + tmp_23*m33))
	dst[11] = d * ((tmp_17*m03 + tmp_20*m13 + tmp_23*m23) -
		(tmp_16*m03 + tmp_21*m13 + tmp_22*m23))
	dst[12] = d * ((tmp_14*m22 + tmp_17*m32 + tmp_13*m12) -
		(tmp_16*m32 + tmp_12*m12 + tmp_15*m22))
	dst[13] = d * ((tmp_20*m32 + tmp_12*m02 + tmp_19*m22) -
		(tmp_18*m22 + tmp_21*m32 + tmp_13*m02))
	dst[14] = d * ((tmp_18*m12 + tmp_23*m32 + tmp_15*m02) -
		(tmp_22*m32 + tmp_14*m02 + tmp_19*m12))
	dst[15] = d * ((tmp_22*m22 + tmp_16*m02 + tmp_21*m12) -
		(tmp_20*m12 + tmp_23*m22 + tmp_17*m02))

	return dst
}

// Perspective returns a perspective matrix
func Perspective(fov, aspect, near, far float32) Matrix4 {
	f := float32(math.Tan((math.Pi * 0.5) - (0.5 * float64(fov))))
	rangeInv := 1.0 / (near - far)
	return Matrix4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (near + far) * rangeInv, -1,
		0, 0, near * far * rangeInv * 2, 0,
	}
}

// LookAt returns a lookat matrix
func LookAt(cameraPosition, target, up Coord) Matrix4 {
	var dst Matrix4
	var zAxis = cameraPosition.Sub(target).Normalize()
	var xAxis = up.Cross(zAxis).Normalize()
	var yAxis = zAxis.Cross(xAxis).Normalize()

	dst[0] = xAxis.X
	dst[1] = xAxis.Y
	dst[2] = xAxis.Z
	dst[3] = 0
	dst[4] = yAxis.X
	dst[5] = yAxis.Y
	dst[6] = yAxis.Z
	dst[7] = 0
	dst[8] = zAxis.X
	dst[9] = zAxis.Y
	dst[10] = zAxis.Z
	dst[11] = 0
	dst[12] = cameraPosition.X
	dst[13] = cameraPosition.Y
	dst[14] = cameraPosition.Z
	dst[15] = 1

	return dst
}

// Identity returns the Identity matrix
func Identity() Matrix4 {
	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// ZTo returns a Z to matrix
func ZTo(z float32) Matrix4 {
	return Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

// Projection returns a Projection matrix
func Projection(width, height, depth float32) Matrix4 {
	return Matrix4{
		2 / width, 0, 0, 0,
		0, -2 / height, 0, 0,
		0, 0, 2 / depth, 0,
		-1, 1, 0, 1,
	}
}

// Translate returns a translating matrix
func Translate(x, y, z float32) Matrix4 {
	return Matrix4{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}
}

// Scale returns a scaling matrix
func Scale(x, y, z float32) Matrix4 {
	return Matrix4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

// Orthographic returns a projection matrix
func Orthographic(left, right, bottom, top, near, far float32) Matrix4 {
	return Matrix4{
		2.0 / (right - left), 0, 0, 0,
		0, 2 / (top - bottom), 0, 0,
		0, 0, 2 / (near - far), 0,
		(left + right) / (left - right), (bottom + top) / (bottom - top), (near + far) / (near - far), 1,
	}
}
