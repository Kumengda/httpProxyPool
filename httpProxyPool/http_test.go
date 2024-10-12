package httpProxyPool

import (
	"fmt"
	"testing"
)

func TestAAA(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(generateRandomHashCode())
	}
}
