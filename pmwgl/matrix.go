package pmwgl

// Coord represents a coordinate
type Coord struct {
	X float32
	Y float32
	Z float32
	W float32
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
func (m Matrix4) MultiplyCoord(c *Coord) Coord {
	return Coord{
		X: (m[ma] * c.X) + (m[mb] * c.Y) + (m[mc] * c.Z) + (m[md] * c.W),
		Y: (m[me] * c.X) + (m[mf] * c.Y) + (m[mg] * c.Z) + (m[mh] * c.W),
		Z: (m[mi] * c.X) + (m[mj] * c.Y) + (m[mk] * c.Z) + (m[ml] * c.W),
		W: (m[ma] * c.X) + (m[mn] * c.Y) + (m[mo] * c.Z) + (m[mp] * c.W),
	}
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
