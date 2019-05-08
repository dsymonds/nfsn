package nfsn

import (
	"testing"
	"time"
)

func TestGenAuthHeader(t *testing.T) {
	// https://members.nearlyfreespeech.net/wiki/API/Introduction : Authentication
	got := genAuthHeader("testuser", time.Unix(1012121212, 0), []byte("dkwo28Sile4jdXkw"),
		"p3kxmRKf9dk3l6ls", "/site/example/getInfo", nil)
	const want = "testuser;1012121212;dkwo28Sile4jdXkw;0fa8932e122d56e2f6d1550f9aab39c4aef8bfc4"
	if got != want {
		t.Errorf("genAuthHeader wrong\n got %q\nwant %q", got, want)
	}
}
