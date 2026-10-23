package main

import (
	"path/filepath"
	"testing"
)

func aggregates(n *Node) (count int) {
	for _, c := range n.Children {
		if c.Path == "" && c.Name == "smaller items" {
			count++
		}
	}
	return
}

func TestSnapshotRoundTripAndCompact(t *testing.T) {
	tmp := t.TempDir()
	cacheDir = func() string { return tmp }
	dir := t.TempDir()
	root := &Node{Name: "r", Path: dir, IsDir: true}
	for i := 0; i < keepPerDir+30; i++ {
		root.Children = append(root.Children, &Node{Name: string(rune('a'+i%26)) + string(rune('0'+i/26)), Path: filepath.Join(dir, "f"+string(rune(i))), Size: int64(1000 - i), Items: 1})
		root.Size += int64(1000 - i)
		root.Items++
	}
	c := compact(root)
	if len(c.Children) != keepPerDir+1 || aggregates(c) != 1 {
		t.Fatalf("compact kept %d entries, %d aggregates", len(c.Children), aggregates(c))
	}
	if again := compact(c); len(again.Children) != len(c.Children) || aggregates(again) != 1 {
		t.Fatalf("compacting twice must not add a second aggregate: %d vs %d", len(again.Children), len(c.Children))
	}

	if err := writeSnapshot(dir, 1234, c); err != nil {
		t.Fatal(err)
	}
	got, err := loadSnapshot(dir)
	if err != nil || got.At != 1234 || got.Tree.Size != root.Size || len(got.Tree.Children) != len(c.Children) {
		t.Fatalf("round trip: %+v, %v", got, err)
	}
	if cacheTime(dir) == 0 {
		t.Fatal("cacheTime should see the file")
	}
	if _, err := loadSnapshot(dir + "-other"); err == nil {
		t.Fatal("a root without a snapshot must not load")
	}
}
