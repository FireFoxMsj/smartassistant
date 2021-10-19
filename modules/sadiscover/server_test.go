package sadiscover

import (
	"fmt"
	"testing"
)

func TestGetKeyAndIV(t *testing.T) {
	var saID = []byte("demo-sa")
	data := uint(0)
	for _, b := range saID {
		data = (data << 8) | uint(b)
	}
	fmt.Println(data)
}