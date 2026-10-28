package gittest

import "testing"

// Ready-made repositories for common test situations.

// WithMergeHistory returns a repo on main with six commits:
//
//	main 1          tagged v0.1.0, the root
//	topic 1         on branch topic, off main 1, never merged
//	feature 1, 2    on branch feature, off main 1
//	main 2          on main, off main 1
//	merge feature   HEAD of main, parents main 2 and feature 2
func WithMergeHistory(t testing.TB) *Repo {
	t.Helper()
	r := New(t)
	r.CommitFile("README.md", "hello\n", "main 1")
	r.Tag("v0.1.0", "first release")
	r.Branch("topic")
	r.CommitFile("topic.txt", "topic\n", "topic 1")
	r.Checkout("main")
	r.Branch("feature")
	r.CommitFile("feature.txt", "one\n", "feature 1")
	r.CommitFile("feature.txt", "two\n", "feature 2")
	r.Checkout("main")
	r.CommitFile("main.txt", "main\n", "main 2")
	if err := r.Merge("feature"); err != nil {
		t.Fatal(err)
	}
	return r
}

// WithConflict returns a repo stopped in the middle of a merge, with conflict.txt
// conflicted between main ("ours") and the branch "theirs".
func WithConflict(t testing.TB) *Repo {
	t.Helper()
	r := New(t)
	r.CommitFile("conflict.txt", "base\n", "base")
	r.Branch("theirs")
	r.CommitFile("conflict.txt", "theirs\n", "theirs change")
	r.Checkout("main")
	r.CommitFile("conflict.txt", "ours\n", "ours change")
	if err := r.Merge("theirs"); err == nil {
		t.Fatal("expected merge conflict")
	}
	return r
}

// WithRemote returns a clone of a bare remote that has the given number of commits on main.
// The clone's "origin" points at the returned remote.
func WithRemote(t testing.TB, commits int) (local, remote *Repo) {
	t.Helper()
	remote = NewBare(t)
	seed := New(t)
	seed.Linear(commits)
	seed.AddRemote("origin", remote)
	seed.Git("push", "-q", "origin", "main")
	return Clone(t, remote), remote
}
