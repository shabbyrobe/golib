package unstructured

// Should support json.Number
type AsInt64 interface {
	Int64() (int64, error)
}

// Should support json.Number
type AsFloat64 interface {
	Float64() (float64, error)
}
