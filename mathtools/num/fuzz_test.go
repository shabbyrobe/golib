package num

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
)

type fuzzOp string
type fuzzType string

// This is the equivalent of passing -num.fuzziter=10000 to 'go test':
const fuzzDefaultIterations = 10000

// These ops are all enabled by default. You can instead pass them explicitly
// on the command line like so: '-num.fuzzop=add -num.fuzzop=sub'
//
// If you add a new op, search for the string 'NEWOP' in this file for all the
// places you need to update.
const (
	fuzzAbs              fuzzOp = "abs"
	fuzzAdd              fuzzOp = "add"
	fuzzAnd              fuzzOp = "and"
	fuzzAsFloat64        fuzzOp = "asfloat64"
	fuzzCmp              fuzzOp = "cmp"
	fuzzDec              fuzzOp = "dec"
	fuzzEqual            fuzzOp = "equal"
	fuzzGreaterOrEqualTo fuzzOp = "gte"
	fuzzGreaterThan      fuzzOp = "gt"
	fuzzInc              fuzzOp = "inc"
	fuzzLessOrEqualTo    fuzzOp = "lte"
	fuzzLessThan         fuzzOp = "lt"
	fuzzLsh              fuzzOp = "lsh"
	fuzzMul              fuzzOp = "mul"
	fuzzNeg              fuzzOp = "neg"
	fuzzOr               fuzzOp = "or"
	fuzzQuo              fuzzOp = "quo"
	fuzzQuoRem           fuzzOp = "quorem"
	fuzzRem              fuzzOp = "rem"
	fuzzRsh              fuzzOp = "rsh"
	fuzzSub              fuzzOp = "sub"
	fuzzXor              fuzzOp = "xor"
)

// These types are all enabled by default. You can instead pass them explicitly
// on the command line like so: '-num.fuzztype=u128 -num.fuzztype=i128'
const (
	fuzzTypeU128 fuzzType = "u128"
	fuzzTypeI128 fuzzType = "i128"
)

var allFuzzTypes = []fuzzType{fuzzTypeU128, fuzzTypeI128}

// allFuzzOps are active by default.
//
// NEWOP: Update this list if a NEW op is added otherwise it won't be
// enabled by default.
var allFuzzOps = []fuzzOp{
	fuzzAbs,
	fuzzAdd,
	fuzzAnd,
	fuzzAsFloat64,
	fuzzCmp,
	fuzzDec,
	fuzzEqual,
	fuzzGreaterOrEqualTo,
	fuzzGreaterThan,
	fuzzInc,
	fuzzLessOrEqualTo,
	fuzzLessThan,
	fuzzLsh,
	fuzzMul,
	fuzzNeg,
	fuzzOr,
	fuzzQuo,
	fuzzQuoRem,
	fuzzRem,
	fuzzRsh,
	fuzzSub,
	fuzzXor,
}

// NEWOP: update this interface if a new op is added.
type fuzzOps interface {
	Name() string // Not an op

	Abs() error
	Add() error
	And() error
	AsFloat64() error
	Cmp() error
	Dec() error
	Equal() error
	GreaterOrEqualTo() error
	GreaterThan() error
	Inc() error
	LessOrEqualTo() error
	LessThan() error
	Lsh() error
	Mul() error
	Neg() error
	Or() error
	Quo() error
	QuoRem() error
	Rem() error
	Rsh() error
	Sub() error
	Xor() error
}

// classic rando!
type rando struct {
	operands []*big.Int
	rng      *rand.Rand
}

func (r *rando) Operands() []*big.Int { return r.operands }
func (r *rando) Clear()               { r.operands = r.operands[:0] }

func (r *rando) U128() U128 {
	bits := uint(r.rng.Intn(128))
	if bits <= 64 {
		return U128{lo: r.rng.Uint64() & ((1 << bits) - 1)}
	} else {
		return U128{hi: r.rng.Uint64() & ((1 << bits) - 1), lo: r.rng.Uint64()}
	}
}

func (r *rando) Intn(n int) int {
	v := int(r.rng.Intn(n))
	r.operands = append(r.operands, new(big.Int).SetInt64(int64(v)))
	return v
}

func (r *rando) Uintn(n int) uint {
	v := uint(r.rng.Intn(n))
	r.operands = append(r.operands, new(big.Int).SetUint64(uint64(v)))
	return v
}

func (r *rando) BigU128() *big.Int {
	var v big.Int

	// FIXME: actually profile the distribution of this to make sure it's doing
	// what's expected of it:
	bits := uint(r.rng.Intn(128))
	if bits <= 64 {
		n := r.rng.Uint64() & ((1 << bits) - 1)
		v.SetUint64(n)
		r.operands = append(r.operands, &v)
		return &v

	} else {
		hi := r.rng.Uint64() & ((1 << (bits - 64)) - 1)
		v.SetUint64(hi).
			Lsh(&v, 64).
			Add(&v, new(big.Int).SetUint64(r.rng.Uint64()))
		r.operands = append(r.operands, &v)
		return &v
	}
}

func (r *rando) BigI128() *big.Int {
	var v big.Int

	// FIXME: actually profile the distribution of this to make sure it's doing
	// what's expected of it:
	bits := uint(r.rng.Intn(127))
	neg := r.rng.Intn(2) == 1

	if bits <= 64 {
		n := r.rng.Uint64() & ((1 << bits) - 1)
		v.SetUint64(n)
		if neg {
			v.Neg(&v)
		}
		r.operands = append(r.operands, &v)
		return &v

	} else {
		hi := r.rng.Uint64() & ((1 << (bits - 64)) - 1)
		v.SetUint64(hi).
			Lsh(&v, 64).
			Add(&v, new(big.Int).SetUint64(r.rng.Uint64()))
		if neg {
			v.Neg(&v)
		}
		r.operands = append(r.operands, &v)
		return &v
	}
}

func checkEqualInt(u int, b int) error {
	if u != b {
		return fmt.Errorf("128(%v) != big(%v)", u, b)
	}
	return nil
}

func checkEqualBool(u bool, b bool) error {
	if u != b {
		return fmt.Errorf("128(%v) != big(%v)", u, b)
	}
	return nil
}

func checkEqualU128(u U128, b *big.Int) error {
	if u.String() != b.String() {
		return fmt.Errorf("u128(%s) != big(%s)", u.String(), b.String())
	}
	return nil
}

func checkFloatU128(orig *big.Int, u U128, b *big.Int) error {
	return checkFloatCommon(orig, u.AsBigInt(), u.String(), b)
}

func checkFloatI128(orig *big.Int, i I128, b *big.Int) error {
	return checkFloatCommon(orig, i.AsBigInt(), i.String(), b)
}

func checkFloatCommon(orig, val *big.Int, valstr string, b *big.Int) error {
	diff := new(big.Int).Set(val)
	diff.Sub(diff, b)

	difff := new(big.Float).SetInt(diff)
	if orig.Cmp(big0) == 0 {
		difff.SetInt(orig)
	} else {
		difff.Quo(difff, new(big.Float).SetInt(orig))
	}

	if difff.Abs(difff).Cmp(floatDiffLimit) > 0 {
		return fmt.Errorf("|u128(%s) - big(%s)| = %s, > %s", valstr, b.String(),
			cleanFloatStr(fmt.Sprintf("%.20f", difff)),
			cleanFloatStr(fmt.Sprintf("%.20f", floatDiffLimit)))
	}
	return nil
}

func checkEqualI128(i I128, b *big.Int) error {
	if i.String() != b.String() {
		return fmt.Errorf("i128(%s) != big(%s)", i.String(), b.String())
	}
	return nil
}

type fuzzU128 struct {
	source *rando
}

func (f fuzzU128) Name() string { return "u128" }

func (f fuzzU128) Abs() error {
	return nil // Always succeeds!
}

func (f fuzzU128) Inc() error {
	b1 := f.source.BigU128()
	u1 := accU128FromBigInt(b1)
	rb := new(big.Int).Add(b1, big1)
	ru := u1.Inc()
	if rb.Cmp(wrapBigU128) >= 0 {
		rb = new(big.Int).Sub(rb, wrapBigU128) // simulate overflow
	}
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Dec() error {
	b1 := f.source.BigU128()
	u1 := accU128FromBigInt(b1)
	rb := new(big.Int).Sub(b1, big1)
	if rb.Cmp(big0) < 0 {
		rb = new(big.Int).Add(wrapBigU128, rb) // simulate underflow
	}
	ru := u1.Dec()
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Add() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).Add(b1, b2)
	if rb.Cmp(wrapBigU128) >= 0 {
		rb = new(big.Int).Sub(rb, wrapBigU128) // simulate overflow
	}
	ru := u1.Add(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Sub() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).Sub(b1, b2)
	if rb.Cmp(big0) < 0 {
		rb = new(big.Int).Add(wrapBigU128, rb) // simulate underflow
	}
	ru := u1.Sub(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Mul() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).Mul(b1, b2)
	for rb.Cmp(wrapBigU128) >= 0 {
		rb = rb.And(rb, maxBigU128) // simulate overflow
	}
	ru := u1.Mul(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Quo() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}
	rb := new(big.Int).Quo(b1, b2)
	ru := u1.Quo(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Rem() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}
	rb := new(big.Int).Rem(b1, b2)
	ru := u1.Rem(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) QuoRem() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}

	rbq := new(big.Int).Quo(b1, b2)
	rbr := new(big.Int).Rem(b1, b2)
	ruq, rur := u1.QuoRem(u2)
	if err := checkEqualU128(ruq, rbq); err != nil {
		return err
	}
	if err := checkEqualU128(rur, rbr); err != nil {
		return err
	}
	return nil
}

func (f fuzzU128) Cmp() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualInt(b1.Cmp(b2), u1.Cmp(u2))
}

func (f fuzzU128) Equal() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualBool(b1.Cmp(b2) == 0, u1.Equal(u2))
}

func (f fuzzU128) GreaterThan() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualBool(b1.Cmp(b2) > 0, u1.GreaterThan(u2))
}

func (f fuzzU128) GreaterOrEqualTo() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualBool(b1.Cmp(b2) >= 0, u1.GreaterOrEqualTo(u2))
}

func (f fuzzU128) LessThan() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualBool(b1.Cmp(b2) < 0, u1.LessThan(u2))
}

func (f fuzzU128) LessOrEqualTo() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	return checkEqualBool(b1.Cmp(b2) <= 0, u1.LessOrEqualTo(u2))
}

func (f fuzzU128) And() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).And(b1, b2)
	ru := u1.And(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Or() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).Or(b1, b2)
	ru := u1.Or(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Xor() error {
	b1, b2 := f.source.BigU128(), f.source.BigU128()
	u1, u2 := accU128FromBigInt(b1), accU128FromBigInt(b2)
	rb := new(big.Int).Xor(b1, b2)
	ru := u1.Xor(u2)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Lsh() error {
	b1 := f.source.BigU128()
	by := f.source.Uintn(128)
	u1 := accU128FromBigInt(b1)
	rb := new(big.Int).Lsh(b1, by)
	rb.And(rb, maxBigU128)
	ru := u1.Lsh(by)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Rsh() error {
	b1 := f.source.BigU128()
	by := f.source.Uintn(128)
	u1 := accU128FromBigInt(b1)
	rb := new(big.Int).Rsh(b1, by)
	ru := u1.Rsh(by)
	return checkEqualU128(ru, rb)
}

func (f fuzzU128) Neg() error {
	return nil // nothing to do here
}

func (f fuzzU128) AsFloat64() error {
	b1 := f.source.BigU128()
	u1 := accU128FromBigInt(b1)
	bf := new(big.Float).SetInt(b1)
	rbf, _ := bf.Float64()
	ruf := u1.AsFloat64()
	rb, _ := new(big.Float).SetFloat64(rbf).Int(new(big.Int))
	ru := U128FromFloat64(ruf)
	return checkFloatU128(b1, ru, rb)
}

type fuzzI128 struct {
	source *rando
}

func (f fuzzI128) Name() string { return "i128" }

func (f fuzzI128) Abs() error {
	b1 := f.source.BigI128()
	i1 := accI128FromBigInt(b1)
	rb := new(big.Int).Abs(b1)
	ru := i1.Abs()
	if rb.Cmp(maxBigI128) > 0 { // overflow is possible if you abs minBig128
		rb = new(big.Int).Add(wrapBigU128, rb)
	}
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Inc() error {
	b1 := f.source.BigI128()
	u1 := accI128FromBigInt(b1)
	rb := new(big.Int).Add(b1, big1)
	ru := u1.Inc()
	if rb.Cmp(maxBigI128) > 0 {
		rb = new(big.Int).Sub(rb, wrapBigU128) // simulate overflow
	}
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Dec() error {
	b1 := f.source.BigI128()
	u1 := accI128FromBigInt(b1)
	rb := new(big.Int).Sub(b1, big1)
	if rb.Cmp(minBigI128) < 0 {
		rb = new(big.Int).Add(wrapBigU128, rb) // simulate underflow
	}
	ru := u1.Dec()
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Add() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	rb := new(big.Int).Add(b1, b2)
	if rb.Cmp(maxBigI128) > 0 {
		rb = new(big.Int).Sub(rb, wrapBigU128) // simulate overflow
	}
	ru := u1.Add(u2)
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Sub() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	rb := new(big.Int).Sub(b1, b2)
	if rb.Cmp(minBigI128) < 0 {
		rb = new(big.Int).Add(wrapBigU128, rb) // simulate underflow
	}
	ru := u1.Sub(u2)
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Mul() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	rb := new(big.Int).Mul(b1, b2)

	if rb.Cmp(maxBigI128) > 0 {
		// simulate overflow
		gap := new(big.Int)
		gap.Sub(rb, minBigI128)
		r := new(big.Int).Rem(gap, wrapBigU128)
		rb = r.Add(r, minBigI128)
	} else if rb.Cmp(minBigI128) < 0 {
		// simulate underflow
		gap := new(big.Int).Set(rb)
		gap.Sub(maxBigI128, gap)
		r := new(big.Int).Rem(gap, wrapBigU128)
		rb = r.Sub(maxBigI128, r)
	}

	ru := u1.Mul(u2)
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Quo() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}
	rb := new(big.Int).Quo(b1, b2)
	ru := u1.Quo(u2)
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) Rem() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}
	rb := new(big.Int).Rem(b1, b2)
	ru := u1.Rem(u2)
	return checkEqualI128(ru, rb)
}

func (f fuzzI128) QuoRem() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	if b2.Cmp(big0) == 0 {
		return nil // Just skip this iteration, we know what happens!
	}

	rbq := new(big.Int).Quo(b1, b2)
	rbr := new(big.Int).Rem(b1, b2)
	ruq, rur := u1.QuoRem(u2)
	if err := checkEqualI128(ruq, rbq); err != nil {
		return err
	}
	if err := checkEqualI128(rur, rbr); err != nil {
		return err
	}
	return nil
}

func (f fuzzI128) Cmp() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualInt(u1.Cmp(u2), b1.Cmp(b2))
}

func (f fuzzI128) Equal() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualBool(u1.Equal(u2), b1.Cmp(b2) == 0)
}

func (f fuzzI128) GreaterThan() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualBool(u1.GreaterThan(u2), b1.Cmp(b2) > 0)
}

func (f fuzzI128) GreaterOrEqualTo() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualBool(u1.GreaterOrEqualTo(u2), b1.Cmp(b2) >= 0)
}

func (f fuzzI128) LessThan() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualBool(u1.LessThan(u2), b1.Cmp(b2) < 0)
}

func (f fuzzI128) LessOrEqualTo() error {
	b1, b2 := f.source.BigI128(), f.source.BigI128()
	u1, u2 := accI128FromBigInt(b1), accI128FromBigInt(b2)
	return checkEqualBool(u1.LessOrEqualTo(u2), b1.Cmp(b2) <= 0)
}

func (f fuzzI128) AsFloat64() error {
	b1 := f.source.BigI128()
	i1 := accI128FromBigInt(b1)
	bf := new(big.Float).SetInt(b1)
	rbf, _ := bf.Float64()
	rif := i1.AsFloat64()
	rb, _ := new(big.Float).SetFloat64(rbf).Int(new(big.Int))
	ri := I128FromFloat64(rif)
	return checkFloatI128(b1, ri, rb)
}

// Bitwise operations on I128 are not supported:
func (f fuzzI128) And() error { return nil }
func (f fuzzI128) Or() error  { return nil }
func (f fuzzI128) Xor() error { return nil }
func (f fuzzI128) Lsh() error { return nil }
func (f fuzzI128) Rsh() error { return nil }

func (f fuzzI128) Neg() error {
	b1 := f.source.BigI128()
	u1 := accI128FromBigInt(b1)
	rb := new(big.Int).Neg(b1)
	if rb.Cmp(maxBigI128) > 0 { // overflow is possible if you negate minBig128
		rb = new(big.Int).Add(wrapBigU128, rb)
	}
	ru := u1.Neg()
	return checkEqualI128(ru, rb)
}

func TestFuzz(t *testing.T) {
	// fuzzOpsActive comes from the -num.fuzzop flag, in TestMain:
	var runFuzzOps = fuzzOpsActive

	// fuzzTypesActive comes from the -num.fuzzop flag, in TestMain:
	var runFuzzTypes = fuzzTypesActive

	var source = &rando{rng: rand.New(rand.NewSource(fuzzSeed))} // Classic rando!
	var totalFailures int

	var fuzzTypes []fuzzOps

	for _, fuzzType := range runFuzzTypes {
		switch fuzzType {
		case fuzzTypeU128:
			fuzzTypes = append(fuzzTypes, &fuzzU128{source: source})
		case fuzzTypeI128:
			fuzzTypes = append(fuzzTypes, &fuzzI128{source: source})
		default:
			panic("unknown fuzz type")
		}
	}

	for _, fuzzImpl := range fuzzTypes {
		var failures = make([]int, len(runFuzzOps))

		for opIdx, op := range runFuzzOps {
			for i := 0; i < fuzzIterations; i++ {
				source.Clear()

				var err error

				// NEWOP: add a new branch here in alphabetical order if a new
				// op is added.
				switch op {
				case fuzzAbs:
					err = fuzzImpl.Abs()
				case fuzzAdd:
					err = fuzzImpl.Add()
				case fuzzAnd:
					err = fuzzImpl.And()
				case fuzzAsFloat64:
					err = fuzzImpl.AsFloat64()
				case fuzzCmp:
					err = fuzzImpl.Cmp()
				case fuzzDec:
					err = fuzzImpl.Dec()
				case fuzzEqual:
					err = fuzzImpl.Equal()
				case fuzzGreaterOrEqualTo:
					err = fuzzImpl.GreaterOrEqualTo()
				case fuzzGreaterThan:
					err = fuzzImpl.GreaterThan()
				case fuzzInc:
					err = fuzzImpl.Inc()
				case fuzzLessOrEqualTo:
					err = fuzzImpl.LessOrEqualTo()
				case fuzzLessThan:
					err = fuzzImpl.LessThan()
				case fuzzLsh:
					err = fuzzImpl.Lsh()
				case fuzzMul:
					err = fuzzImpl.Mul()
				case fuzzNeg:
					err = fuzzImpl.Neg()
				case fuzzOr:
					err = fuzzImpl.Or()
				case fuzzQuo:
					err = fuzzImpl.Quo()
				case fuzzQuoRem:
					err = fuzzImpl.QuoRem()
				case fuzzRem:
					err = fuzzImpl.Rem()
				case fuzzRsh:
					err = fuzzImpl.Rsh()
				case fuzzSub:
					err = fuzzImpl.Sub()
				case fuzzXor:
					err = fuzzImpl.Xor()
				default:
					panic(fmt.Errorf("unsupported op %q", op))
				}

				if err != nil {
					failures[opIdx]++
					t.Logf("%s: %s\n", op.Print(source.Operands()...), err)
				}
			}
		}

		for opIdx, cnt := range failures {
			if cnt > 0 {
				totalFailures += cnt
				t.Logf("impl %s, op %s: %d/%d failed", fuzzImpl.Name(), string(runFuzzOps[opIdx]), cnt, fuzzIterations)
			}
		}
	}

	if totalFailures > 0 {
		t.Fail()
	}
}

func (op fuzzOp) Print(operands ...*big.Int) string {
	// NEWOP: please add a human-readale format for your op here; this is used
	// for reporting errors and should show the operation, i.e. "2 + 2".
	//
	// It should be safe to assume the appropriate number of operands are set
	// in 'operands'; if not, it's a bug to be fixed elsewhere.
	switch op {
	case fuzzAsFloat64:
		return fmt.Sprintf("float64(%d)", operands[0])

	case fuzzInc, fuzzDec:
		return fmt.Sprintf("%d%s", operands[0], op.String())

	case fuzzNeg:
		return fmt.Sprintf("-%d", operands[0])

	case fuzzAbs:
		return fmt.Sprintf("|%d|", operands[0])

	case fuzzAdd, fuzzSub, fuzzCmp, fuzzEqual, fuzzGreaterThan, fuzzGreaterOrEqualTo,
		fuzzLessThan, fuzzLessOrEqualTo, fuzzAnd, fuzzOr, fuzzXor, fuzzLsh, fuzzRsh,
		fuzzMul, fuzzQuo, fuzzRem, fuzzQuoRem: // simple binary case:
		return fmt.Sprintf("%d %s %d", operands[0], op.String(), operands[1])

	default:
		return string(op)
	}
}

func (op fuzzOp) String() string {
	// NEWOP: please add a short string representation of this op, as if
	// the operands were in a sum.
	switch op {
	case fuzzAbs:
		return "|x|"
	case fuzzAdd:
		return "+"
	case fuzzAnd:
		return "&"
	case fuzzAsFloat64:
		return "float64()"
	case fuzzCmp:
		return "<=>"
	case fuzzDec:
		return "--"
	case fuzzEqual:
		return "=="
	case fuzzGreaterThan:
		return ">"
	case fuzzGreaterOrEqualTo:
		return ">="
	case fuzzInc:
		return "++"
	case fuzzLessThan:
		return "<"
	case fuzzLessOrEqualTo:
		return "<="
	case fuzzLsh:
		return "<<"
	case fuzzMul:
		return "*"
	case fuzzNeg:
		return "-"
	case fuzzOr:
		return "|"
	case fuzzQuo:
		return "/"
	case fuzzQuoRem:
		return "/%"
	case fuzzRem:
		return "%"
	case fuzzRsh:
		return ">>"
	case fuzzSub:
		return "-"
	case fuzzXor:
		return "^"
	default:
		return string(op)
	}
}
