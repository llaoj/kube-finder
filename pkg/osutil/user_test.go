package osutil

import (
	"fmt"
	"testing"
)

func TestLookupUserIdFrom(t *testing.T) {
	u, err := LookupUserIdFrom("/Users/weiyangwang/go/src/github.com/llaoj/kube-finder/passwd", fmt.Sprint(998))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(u.Username)
}
