package jsonwriter

type node struct {
	children int

	kind nodeKind
}

type nodeKind int

func (n nodeKind) String() string {
	names := ""
	v := 0
	for i := nodeKind(1); i <= keyState; i <<= 1 {
		set := n & i
		cur := ""
		switch set {
		case rootNode:
			cur = "root"
		case listNode:
			cur = "list"
		case objectNode:
			cur = "object"
		case valueState:
			cur = ""
		case keyState:
			cur = "key"
		}
		if cur != "" {
			if v > 1 {
				names += ", " + cur
			} else {
				names = cur
			}
			v++
		}
	}
	return names
}

func (n nodeKind) StateString() string {
	if n&valueState != 0 {
		return "value"
	} else {
		return "key"
	}
}

const (
	rootNode nodeKind = 1 << iota
	objectNode
	listNode

	valueState
	keyState
)
