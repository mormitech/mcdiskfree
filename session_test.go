package main

import "testing"

func TestSessionRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	cacheDir = func() string { return tmp }
	a := NewApp(nil)

	if s := a.LoadSession(); s.Root != "" {
		t.Fatalf("fresh cache dir must give an empty session, got %+v", s)
	}
	a.SaveSession(tmp, tmp+"/sub")
	if s := a.LoadSession(); s.Root != tmp || s.Path != tmp+"/sub" {
		t.Fatalf("round trip gave %+v", s)
	}
	a.SaveSession("", "") // the disk list was on screen last
	if s := a.LoadSession(); s.Root != "" {
		t.Fatalf("the disk list must restore as an empty session, got %+v", s)
	}
	a.SaveSession(tmp, tmp)
	a.SaveSession("", "")
	if s := a.LoadSession(); s.Root != "" {
		t.Fatalf("leaving for the disk list must replace the saved disk, got %+v", s)
	}
	a.SaveSession(tmp+"/gone", tmp+"/gone")
	if s := a.LoadSession(); s.Root != "" {
		t.Fatalf("a root that no longer exists must not be restored, got %+v", s)
	}
}
