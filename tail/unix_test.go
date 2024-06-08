package tail

import (
	"fmt"
	"os"
	"testing"
)

func TestTail(t *testing.T) {
	temp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	u, err := NewUnix(temp.Name())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for i := 0; i < 5; i++ {
			if _, err := temp.WriteString(fmt.Sprintf("line %d\n", i)); err != nil {
				panic(err)
			}
		}
	}()

	lines := []string{}
	for i := 0; i < 5; i++ {
		if !u.Scan() {
			t.Fatal(i)
		}
		lines = append(lines, u.Line())
	}

	fmt.Println(lines)
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}
}
