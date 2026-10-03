package memory

import (
	"testing"

	"github.com/JIeeiroSst/webrtc-service/internal/domain"
)

type stubPeer struct{ id, room string }

func (p *stubPeer) ID() string                 { return p.id }
func (p *stubPeer) RoomID() string             { return p.room }
func (p *stubPeer) Send(domain.Message) error  { return nil }
func (p *stubPeer) Media() domain.MediaState   { return domain.MediaState{} }
func (p *stubPeer) SetMedia(domain.MediaState) {}
func (p *stubPeer) Close()                     {}

func TestAddReturnsOtherMembers(t *testing.T) {
	r := NewRoomRegistry()
	a := &stubPeer{"a", "room"}
	b := &stubPeer{"b", "room"}
	other := &stubPeer{"c", "elsewhere"}

	r.Add(a, 0)
	r.Add(other, 0)
	replaced, others, _ := r.Add(b, 0)

	if replaced != nil {
		t.Fatalf("replaced = %v, want nil", replaced)
	}
	if len(others) != 1 || others[0] != a {
		t.Fatalf("others = %v, want [a]", others)
	}
}

func TestAddReplacesSameID(t *testing.T) {
	r := NewRoomRegistry()
	old := &stubPeer{"a", "room"}
	b := &stubPeer{"b", "room"}
	fresh := &stubPeer{"a", "room"}

	r.Add(old, 0)
	r.Add(b, 0)
	replaced, others, _ := r.Add(fresh, 0)

	if replaced != old {
		t.Fatalf("replaced = %v, want old session", replaced)
	}
	if len(others) != 1 || others[0] != b {
		t.Fatalf("others = %v, want [b]", others)
	}
	if got, _ := r.Get("room", "a"); got != fresh {
		t.Fatalf("Get = %v, want fresh session", got)
	}
}

func TestRemoveIgnoresReplacedSession(t *testing.T) {
	r := NewRoomRegistry()
	old := &stubPeer{"a", "room"}
	fresh := &stubPeer{"a", "room"}
	r.Add(old, 0)
	r.Add(fresh, 0)

	if r.Remove(old) {
		t.Fatal("Remove(old) = true, want false")
	}
	if _, ok := r.Get("room", "a"); !ok {
		t.Fatal("fresh session was removed by its predecessor")
	}
	if !r.Remove(fresh) {
		t.Fatal("Remove(fresh) = false, want true")
	}
}

func TestRemoveDropsEmptyRoom(t *testing.T) {
	r := NewRoomRegistry().(*RoomRegistry)
	a := &stubPeer{"a", "room"}
	r.Add(a, 0)
	r.Remove(a)

	if _, ok := r.rooms["room"]; ok {
		t.Fatal("empty room was kept")
	}
	if got := r.Members("room"); len(got) != 0 {
		t.Fatalf("Members = %v, want none", got)
	}
}

func TestAddEnforcesCapacity(t *testing.T) {
	r := NewRoomRegistry()
	r.Add(&stubPeer{"a", "room"}, 2)
	r.Add(&stubPeer{"b", "room"}, 2)

	if _, _, err := r.Add(&stubPeer{"c", "room"}, 2); err != domain.ErrRoomFull {
		t.Fatalf("third peer err = %v, want ErrRoomFull", err)
	}

	if _, _, err := r.Add(&stubPeer{"a", "room"}, 2); err != nil {
		t.Fatalf("replacement err = %v, want nil", err)
	}
	if rooms, peers := r.Stats(); rooms != 1 || peers != 2 {
		t.Fatalf("Stats = %d rooms, %d peers; want 1, 2", rooms, peers)
	}
}
