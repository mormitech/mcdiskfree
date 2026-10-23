package disk

import "testing"

func TestVolumesAndUsage(t *testing.T) {
	vols := Volumes()
	if len(vols) == 0 {
		t.Fatal("no volumes found")
	}
	for _, v := range vols {
		if v.Total == 0 || v.Free > v.Total {
			t.Errorf("odd volume %+v", v)
		}
	}
	u, err := UsageOf(vols[0].Path)
	if err != nil || u.Total == 0 {
		t.Fatalf("UsageOf(%q) = %+v, %v", vols[0].Path, u, err)
	}
	t.Logf("volumes: %+v", vols)
}
