package app

import (
	"testing"
	"time"

	"github.com/risadams/gitrevenge/internal/gittest"
)

func TestGraphRows(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.WithMergeHistory(t)
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	var page GraphPage
	deadline := time.Now().Add(10 * time.Second)
	for {
		if page, err = a.GraphRows(tab.Path, 0, 50); err != nil {
			t.Fatal(err)
		}
		if !page.Loading || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if page.Total != 6 || len(page.Rows) != 6 || page.Rows[0].Subject == "" {
		t.Fatalf("page = total %d, rows %d, loading %v", page.Total, len(page.Rows), page.Loading)
	}
	if page, _ := a.GraphRows(tab.Path, 4, 50); len(page.Rows) != 2 || page.Rows[0].Index != 4 {
		t.Fatalf("offset page = %+v", page.Rows)
	}
	if _, err := a.GraphRows("C:/not/open", 0, 10); err == nil {
		t.Fatal("expected error for a repo that isn't open")
	}
}

func TestGraphRefsAndCommitRow(t *testing.T) {
	a := newTestApp(t)
	fx := gittest.WithMergeHistory(t)
	tab, err := a.OpenRepo(fx.Dir)
	if err != nil {
		t.Fatal(err)
	}
	var page GraphPage
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if page, _ = a.GraphRows(tab.Path, 0, 50); !page.Loading {
			break
		}
	}
	head := fx.Head()
	labels := page.Refs[head]
	if len(labels) != 1 || labels[0].Short != "main" || !labels[0].Head {
		t.Fatalf("labels on HEAD = %+v", labels)
	}
	refs, err := a.Refs(tab.Path)
	if err != nil || len(refs) != 4 { // main, topic, feature, v0.1.0
		t.Fatalf("refs = %+v, %v", refs, err)
	}
	if row, _ := a.CommitRow(tab.Path, head); row != 0 {
		t.Fatalf("CommitRow(HEAD) = %d, want 0", row)
	}
	if row, _ := a.CommitRow(tab.Path, "0000000000000000000000000000000000000000"); row != -1 {
		t.Fatalf("unknown commit row = %d", row)
	}
}
