package devicelimit

import (
	"testing"

	M "github.com/sagernet/sing/common/metadata"
)

func TestLimiterCountsDistinctSourceAddresses(t *testing.T) {
	limiter := &Limiter{users: map[string]map[string]int{}}
	source1 := M.ParseSocksaddr("192.0.2.1:1234")
	source2 := M.ParseSocksaddr("192.0.2.2:1234")
	source3 := M.ParseSocksaddr("192.0.2.3:1234")

	release1, err := limiter.Acquire("alice", 2, source1)
	if err != nil {
		t.Fatal(err)
	}
	release1Again, err := limiter.Acquire("alice", 2, M.ParseSocksaddr("192.0.2.1:5678"))
	if err != nil {
		t.Fatal(err)
	}
	release2, err := limiter.Acquire("alice", 2, source2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = limiter.Acquire("alice", 2, source3); err == nil {
		t.Fatal("expected third distinct source to be rejected")
	}

	release1()
	if _, err = limiter.Acquire("alice", 2, source3); err == nil {
		t.Fatal("same source should still hold one reference")
	}
	release1Again()
	release3, err := limiter.Acquire("alice", 2, source3)
	if err != nil {
		t.Fatal(err)
	}
	release2()
	release3()
}

func TestLimiterIgnoresUnlimitedUsers(t *testing.T) {
	limiter := &Limiter{users: map[string]map[string]int{}}
	for _, source := range []M.Socksaddr{
		M.ParseSocksaddr("192.0.2.1:1234"),
		M.ParseSocksaddr("192.0.2.2:1234"),
		M.ParseSocksaddr("192.0.2.3:1234"),
	} {
		release, err := limiter.Acquire("alice", 0, source)
		if err != nil {
			t.Fatal(err)
		}
		if release != nil {
			t.Fatal("unlimited users should not allocate a release function")
		}
	}
}
