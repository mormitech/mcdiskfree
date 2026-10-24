package main

import (
	"strings"
	"testing"
)

func TestDemoFlag(t *testing.T) {
	for args, want := range map[string]string{"--mock=windows": "windows", "--mock linux": "linux", "-demo macos": "macos", "": ""} {
		if got, _ := startFlags(strings.Fields(args)); got != want {
			t.Errorf("startFlags(%q) = %q, want %q", args, got, want)
		}
	}
	if got, about := startFlags([]string{"-psn_0_12345"}); got != "" || about {
		t.Errorf("unknown flags must be ignored, got %q", got)
	}
}

func TestAboutFlag(t *testing.T) {
	if _, about := startFlags([]string{"--about"}); !about {
		t.Error("--about not read")
	}
	if _, about := startFlags(nil); about {
		t.Error("about must be off by default")
	}
}

func TestMockDisks(t *testing.T) {
	for _, p := range []string{"windows", "linux", "macos"} {
		m, err := loadMock(p)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.volumes) < 2 {
			t.Fatalf("%s: want several disks", p)
		}
		for _, v := range m.volumes {
			root := m.roots[v.Path]
			if root == nil || root.Size <= 0 || root.Items < 10 {
				t.Fatalf("%s %s: tree looks empty", p, v.Name)
			}
			if root.Size > int64(v.Total-v.Free) {
				t.Errorf("%s %s: files (%d) exceed used space (%d)", p, v.Name, root.Size, v.Total-v.Free)
			}
			if m.index[root.Children[0].Path] != root.Children[0] {
				t.Errorf("%s %s: index is incomplete", p, v.Name)
			}
		}
	}
	if _, err := loadMock("beos"); err == nil {
		t.Error("unknown platform must fail")
	}
}

func TestMockTrash(t *testing.T) {
	m, _ := loadMock("windows")
	a := NewApp(m)
	v := m.volumes[0]
	if a.CachedScan(v.Path) == nil {
		t.Fatal("no saved scan for a mock disk")
	}
	root := m.roots[v.Path]
	victim := root.Children[0]
	before, free := root.Size, m.volumes[0].Free
	res := a.Trash([]string{victim.Path, v.Path})
	if len(res.Trashed) != 1 || len(res.Failed) != 1 {
		t.Fatalf("trash result %+v", res)
	}
	if root.Size != before-victim.Size || m.volumes[0].Free != free+uint64(victim.Size) {
		t.Error("sizes were not corrected")
	}
	if _, err := a.View(victim.Path); err == nil {
		t.Error("trashed folder is still in the tree")
	}
}
