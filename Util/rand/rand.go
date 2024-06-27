//伪随机算法（梅森旋转）
package rand

type Random struct {
	index int
	mt    [624]uint32
}

func UInt32(x int64) uint32 {
	return uint32(0xFFFFFFFF & x)
}

func (random *Random) Init(seed uint32) {
	random.index = 624
	random.mt[0] = seed
	for i := 1; i < 624; i++ {
		random.mt[i] = UInt32(1812433253*(int64(random.mt[i-1])^int64(random.mt[i-1])>>30) + int64(i))
	}
}

func (random *Random) twist() {
	for i := 0; i < 624; i++ {
		y := UInt32((int64(random.mt[i]) & 0x80000000) + (int64(random.mt[(i+1)%624]) & 0x7fffffff))
		random.mt[i] = random.mt[(i+397)%624] ^ y>>1
		if y%2 != 0 {
			random.mt[i] = random.mt[i] ^ 0x9908b0df
		}
	}

	random.index = 0
}

func (random *Random) Next() uint32 {
	if random.index >= 624 {
		random.twist()
	}

	y := int64(random.mt[random.index])
	//Right shift by 11 bits
	y = y ^ y>>11
	//Shift y left by 7 and take the bitwise and of 2636928640
	y = y ^ y<<7&2636928640
	//Shift y left by 15 and take the bitwise and of y and 4022730752
	y = y ^ y<<15&4022730752
	//Right shift by 18 bits
	y = y ^ y>>18

	random.index = random.index + 1
	return UInt32(y)
}

func (random *Random) Rand(max int) int {
	rand := random.Next()
	return int(rand%uint32(max)) + 1
}
