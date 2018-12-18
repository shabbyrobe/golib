package num

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"regexp"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

var u64 = U128From64

func u128s(s string) U128 {
	b, err := U128FromString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func randU128(scratch []byte) U128 {
	rand.Read(scratch)
	u := U128{}
	u.lo = binary.LittleEndian.Uint64(scratch)

	if scratch[0]%2 == 1 {
		// if we always generate hi bits, the universe will die before we
		// test a number < maxInt64
		u.hi = binary.LittleEndian.Uint64(scratch[8:])
	}
	return u
}

func TestU128FromSize(t *testing.T) {
	tt := assert.WrapTB(t)
	tt.MustEqual(U128From8(255), u128s("255"))
	tt.MustEqual(U128From16(65535), u128s("65535"))
	tt.MustEqual(U128From32(4294967295), u128s("4294967295"))
}

func TestU128Add(t *testing.T) {
	for _, tc := range []struct {
		a, b, c U128
	}{
		{u64(1), u64(2), u64(3)},
		{u64(10), u64(3), u64(13)},
		{MaxU128, u64(1), u64(0)},                               // Overflow wraps
		{u64(maxUint64), u64(1), u128s("18446744073709551616")}, // lo carries to hi
		{u128s("18446744073709551615"), u128s("18446744073709551615"), u128s("36893488147419103230")},
	} {
		t.Run(fmt.Sprintf("%s+%s=%s", tc.a, tc.b, tc.c), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustAssert(tc.c.Equal(tc.a.Add(tc.b)))
		})
	}
}

func TestU128Inc(t *testing.T) {
	for _, tc := range []struct {
		a, b U128
	}{
		{u64(1), u64(2)},
		{u64(10), u64(11)},
		{u64(maxUint64), u128s("18446744073709551616")},
		{u64(maxUint64), u64(maxUint64).Add(u64(1))},
		{MaxU128, u64(0)},
	} {
		t.Run(fmt.Sprintf("%s+1=%s", tc.a, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			inc := tc.a.Inc()
			tt.MustAssert(tc.b.Equal(inc), "%s + 1 != %s, found %s", tc.a, tc.b, inc)
		})
	}
}

func TestU128Dec(t *testing.T) {
	for _, tc := range []struct {
		a, b U128
	}{
		{u64(1), u64(0)},
		{u64(10), u64(9)},
		{u64(maxUint64), u128s("18446744073709551614")},
		{u64(0), MaxU128},
		{u64(maxUint64).Add(u64(1)), u64(maxUint64)},
	} {
		t.Run(fmt.Sprintf("%s-1=%s", tc.a, tc.b), func(t *testing.T) {
			tt := assert.WrapTB(t)
			dec := tc.a.Dec()
			tt.MustAssert(tc.b.Equal(dec), "%s - 1 != %s, found %s", tc.a, tc.b, dec)
		})
	}
}

func TestU128Mul(t *testing.T) {
	tt := assert.WrapTB(t)

	u := U128From64(maxUint64)
	v := u.Mul(U128From64(maxUint64))

	var v1, v2 big.Int
	v1.SetUint64(maxUint64)
	v2.SetUint64(maxUint64)
	tt.MustEqual(v.String(), v1.Mul(&v1, &v2).String())
}

func TestU128Div(t *testing.T) {
	for _, tc := range []struct {
		u, by, q, r U128
	}{
		{u: u64(1), by: u64(2), q: u64(0), r: u64(1)},
		{u: u64(10), by: u64(3), q: u64(3), r: u64(1)},

		// These test cases were found by the fuzzer and exposed a bug in the 128-bit divisor
		// branch of divmod128by128:
		// 3289699161974853443944280720275488 / 9261249991223143249760: u128(48100516172305203) != big(355211139435)
		// 51044189592896282646990963682604803 / 15356086376658915618524: u128(16290274193854465) != big(3324036368438)
		// 555579170280843546177 / 21475569273528505412: u128(12) != big(25)
		// 2949247824660989947922443980 / 302683126529761752263: u128(455939057376533098) != big(9743680)
		// 971472458247603477981631204590 / 350420: u128(2772305826605587383713792) != big(2772308824403868152450291)
		// 11096903380546347857972634826445138634 / 42697940829956011330806: u128(2668015570451510) != big(259893174350951)
		// 2217478155689385033283681835376395 / 40289020034608988764162419152803: u128(8539690) != big(55)
		// 18142638712496683201489267 / 1365: u128(13281655733070877163520) != big(13291310412085482198893)
		// 62103722249961132848385208578 / 12388: u128(5013197849935750004473856) != big(5013216197123113726863513)
		// 1718097515653571439501189267179258857 / 164705935682477192600288: u128(1019612999867505) != big(10431302967524)
		// 3661431417457026177118536052088 / 12365909340431256211712: u128(19939461649052568) != big(296090753)
		// 24298963230542467290270140341245 / 421787975024059225589: u128(341633803428111640) != big(57609426226)
		// 1467621860654665417145484978231305494 / 633124970193674922054836661078187921: u128(1) != big(2)
		// 2802142065843234774980457 / 620: u128(4519452298058840145920) != big(4519583977166507701581)
		// 20224588157632265704346736243251064 / 489601690837183261247311421165: u128(167979940) != big(41308)
		// 23896954878490934294198761890 / 31420116731669012347740838: u128(8434227721305) != big(760)
		// 3918866837393595579648743745051 / 1949638892805710269373538: u128(4355324532566) != big(2010047)
		// 21097982161006329403046439215655211 / 68699402632727790927766640881958: u128(4761867) != big(307)
		// 73935111427887097261995043834539 / 320997938157: u128(221360928884514619392) != big(230328929376877976548)
		// 1851429540072980778899995824609 / 18: u128(102857196665035250408582283264) != big(102857196670721154383333101367)
		// 4711594723937775520752870124446537734 / 820496178458499857061155: u128(277610350355589) != big(5742372539491)
		// 269694753853353588774761351 / 7702535: u128(18446744073709551616) != big(35013765449083138054)
		// 14054597138817478941697630323 / 78953527: u128(166020696663385964544) != big(178011010690092146759)
		// 9050696263923660666457960369273779672 / 93657: u128(96636623679206658807759296790528) != big(96636623679208822260567393459899)
	} {
		t.Run(fmt.Sprintf("%s÷%s=%s,%s", tc.u, tc.by, tc.q, tc.r), func(t *testing.T) {
			tt := assert.WrapTB(t)
			q, r := tc.u.QuoRem(tc.by)
			tt.MustEqual(tc.q.String(), q.String())
			tt.MustEqual(tc.r.String(), r.String())

			uBig := tc.u.AsBigInt()
			byBig := tc.by.AsBigInt()

			qBig, rBig := new(big.Int).Set(uBig), new(big.Int).Set(uBig)
			qBig = qBig.Quo(qBig, byBig)
			rBig = rBig.Rem(rBig, byBig)

			tt.MustEqual(tc.q.String(), qBig.String())
			tt.MustEqual(tc.r.String(), rBig.String())
		})
	}
}

func TestU128AsFloat(t *testing.T) {
	for _, tc := range []struct {
		a   U128
		out string
	}{
		{u128s("2384067163226812360730"), "2384067163226812448768"},
	} {
		t.Run(fmt.Sprintf("float64(%s)=%s", tc.a, tc.out), func(t *testing.T) {
			tt := assert.WrapTB(t)
			tt.MustEqual(tc.out, cleanFloatStr(fmt.Sprintf("%f", tc.a.AsFloat64())))
		})
	}
}

var trimFloatPattern = regexp.MustCompile(`(\.0+$|(\.\d+[1-9])\0+$)`)

func cleanFloatStr(str string) string {
	return trimFloatPattern.ReplaceAllString(str, "$2")
}

func TestU128Rsh(t *testing.T) {
	for _, tc := range []struct {
		u  U128
		by uint
		r  U128
	}{
		{u: u64(2), by: 1, r: u64(1)},
		{u: u64(1), by: 2, r: u64(0)},
		{u: u128s("36893488147419103232"), by: 1, r: u128s("18446744073709551616")}, // (1<<65) - 1

		// These test cases were found by the fuzzer:
		{u: u128s("2465608830469196860151950841431"), by: 104, r: u64(0)},
		{u: u128s("377509308958315595850564"), by: 58, r: u64(1309748)},
		{u: u128s("8504691434450337657905929307096"), by: 74, r: u128s("450234615")},
		{u: u128s("11595557904603123290159404941902684322"), by: 50, r: u128s("10298924295251697538375")},
		{u: u128s("176613673099733424757078556036831904"), by: 75, r: u128s("4674925001596")},
		{u: u128s("3731491383344351937489898072501894878"), by: 112, r: u64(718)},
	} {
		t.Run(fmt.Sprintf("%s>>%d=%s", tc.u, tc.by, tc.r), func(t *testing.T) {
			tt := assert.WrapTB(t)

			ub := tc.u.AsBigInt()
			ub.Rsh(ub, tc.by).And(ub, maxBigU128)

			ru := tc.u.Rsh(tc.by)
			tt.MustEqual(tc.r.String(), ru.String(), "%s != %s; big: %s", tc.r, ru, ub)
			tt.MustEqual(ub.String(), ru.String())
		})
	}
}

func TestU128Lsh(t *testing.T) {
	for _, tc := range []struct {
		u  U128
		by uint
		r  U128
	}{
		{u: u64(2), by: 1, r: u64(4)},
		{u: u64(1), by: 2, r: u64(4)},
		{u: u128s("18446744073709551615"), by: 1, r: u128s("36893488147419103230")}, // (1<<64) - 1

		// These cases were found by the fuzzer:
		{u: u128s("5080864651895"), by: 57, r: u128s("732229764895815899943471677440")},
		{u: u128s("63669103"), by: 85, r: u128s("2463079120908903847397520463364096")},
		{u: u128s("2465608830469196860151950841431"), by: 104, r: u128s("50008488221956801743883323727223890761457150435029728043728896")},
		{u: u128s("377509308958315595850564"), by: 58, r: u128s("108809650121828068156972983880952227823616")},
		{u: u128s("8504691434450337657905929307096"), by: 74, r: u128s("160649079108787355396833790150802372769627880516747264")},
		{u: u128s("11595557904603123290159404941902684322"), by: 50, r: u128s("13055437564580908863505186908351446169053281206140928")},
		{u: u128s("173760885"), by: 68, r: u128s("51285161209860430747989442560")},
		{u: u128s("213"), by: 65, r: u128s("7858312975400268988416")},
		{u: u128s("176613673099733424757078556036831904"), by: 75, r: u128s("6672275922101419229538799409302378069369212160007284457472")},
		{u: u128s("40625"), by: 55, r: u128s("1463669878895411200000")},
	} {
		t.Run(fmt.Sprintf("%s<<%d=%s", tc.u, tc.by, tc.r), func(t *testing.T) {
			tt := assert.WrapTB(t)

			ub := tc.u.AsBigInt()
			ub.Lsh(ub, tc.by).And(ub, maxBigU128)

			ru := tc.u.Lsh(tc.by)
			tt.MustEqual(tc.r.String(), ru.String(), "%s != %s; big: %s", tc.r, ru, ub)
			tt.MustEqual(ub.String(), ru.String())
		})
	}
}

func TestU128Float64Random(t *testing.T) {
	tt := assert.WrapTB(t)

	bts := make([]byte, 16)

	// The percentage of the difference between the input number and the output
	// number relative to the input number after performing the transform
	// U128(float64(U128)) must not be more than this very reasonable limit:
	limit := new(big.Float).SetFloat64(0.00000000000001)

	for i := 0; i < 100000; i++ {
		u := randU128(bts)

		f := u.AsFloat64()
		r := U128FromFloat64(f)
		diff := DifferenceU128(u, r)

		ubig, diffBig := u.AsBigFloat(), diff.AsBigFloat()
		pct := new(big.Float).Quo(diffBig, ubig)

		tt.MustAssert(pct.Cmp(limit) < 0, "%s", pct)
	}
}

func TestU128MarshalJSON(t *testing.T) {
	tt := assert.WrapTB(t)
	bts := make([]byte, 16)

	for i := 0; i < 20000; i++ {
		u := randU128(bts)

		bts, err := json.Marshal(u)
		tt.MustOK(err)

		var result U128
		tt.MustOK(json.Unmarshal(bts, &result))
		tt.MustAssert(result.Equal(u))
	}
}

var (
	BenchUResult        U128
	BenchIntResult      int
	BenchFloatResult    float64
	BenchBigFloatResult *big.Float
)

func BenchmarkU128Mul(b *testing.B) {
	u := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchUResult = u.Mul(u)
	}
}

func BenchmarkU128Add(b *testing.B) {
	u := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchUResult = u.Add(u)
	}
}

func BenchmarkU128QuoRem(b *testing.B) {
	u := U128From64(maxUint64)
	by := U128From64(121525124)
	for i := 0; i < b.N; i++ {
		BenchUResult, _ = u.QuoRem(by)
	}
}

func BenchmarkU128CmpEqual(b *testing.B) {
	u := U128From64(maxUint64)
	n := U128From64(maxUint64)
	for i := 0; i < b.N; i++ {
		BenchIntResult = u.Cmp(n)
	}
}

func BenchmarkU128Lsh(b *testing.B) {
	for _, tc := range []struct {
		in U128
		sh uint
	}{
		{u64(maxUint64), 1},
		{u64(maxUint64), 2},
		{u64(maxUint64), 8},
		{u64(maxUint64), 64},
		{u64(maxUint64), 126},
		{u64(maxUint64), 127},
		{u64(maxUint64), 128},
		{MaxU128, 1},
		{MaxU128, 2},
		{MaxU128, 8},
		{MaxU128, 64},
		{MaxU128, 126},
		{MaxU128, 127},
		{MaxU128, 128},
	} {
		b.Run(fmt.Sprintf("%s>>%d", tc.in, tc.sh), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchUResult = tc.in.Lsh(tc.sh)
			}
		})
	}
}

func BenchmarkU128AsBigFloat(b *testing.B) {
	n := u128s("36893488147419103230")
	for i := 0; i < b.N; i++ {
		BenchBigFloatResult = n.AsBigFloat()
	}
}

func BenchmarkU128AsFloat(b *testing.B) {
	n := u128s("36893488147419103230")
	for i := 0; i < b.N; i++ {
		BenchFloatResult = n.AsFloat64()
	}
}

func BenchmarkU128FromFloat(b *testing.B) {
	for _, pow := range []float64{1, 63, 64, 65, 127, 128} {
		b.Run(fmt.Sprintf("pow%d", int(pow)), func(b *testing.B) {
			f := math.Pow(2, pow)
			for i := 0; i < b.N; i++ {
				BenchUResult = U128FromFloat64(f)
			}
		})
	}
}

func BenchmarkBigIntMul(b *testing.B) {
	var max big.Int
	max.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		var dest big.Int
		dest.Mul(&dest, &max)
	}
}

func BenchmarkBigIntAdd(b *testing.B) {
	var max big.Int
	max.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		var dest big.Int
		dest.Add(&dest, &max)
	}
}

func BenchmarkBigIntDiv(b *testing.B) {
	u := new(big.Int).SetUint64(maxUint64)
	by := new(big.Int).SetUint64(121525124)
	for i := 0; i < b.N; i++ {
		var z big.Int
		z.Div(u, by)
	}
}

func BenchmarkBigIntCmpEqual(b *testing.B) {
	var v1, v2 big.Int
	v1.SetUint64(maxUint64)
	v2.SetUint64(maxUint64)

	for i := 0; i < b.N; i++ {
		BenchIntResult = v1.Cmp(&v2)
	}
}
