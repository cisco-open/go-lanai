package cmdutils

import (
	"context"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"path/filepath"
)

const (
	errTmplGetMarkedTag = `error when getting marked tag %s: %v`
)

type GitFileMatcher func(path string) bool

func GitFilePattern(patterns ...string) GitFileMatcher {
	for i, p := range patterns {
		patterns[i] = filepath.Clean(p)
	}
	return func(path string) bool {
		for _, pattern := range patterns {
			if matched, e := doublestar.Match(pattern, path); e == nil && matched {
				return true
			}
		}
		return false
	}
}

type GitUtils struct {
	ctx context.Context
	repo *git.Repository
}

func NewGitUtilsWithWorkingDir() (*GitUtils, error) {
	return NewGitUtilsWithPath(GlobalArgs.WorkingDir)
}

func NewGitUtilsWithPath(path string) (*GitUtils, error) {
	if path == "" {
		return nil, fmt.Errorf(`cannot open repository: path is not specified`)
	}
	repo, e := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
	})
	if e != nil {
		return nil, e
	}
	return &GitUtils{
		ctx: context.Background(),
		repo: repo,
	}, nil
}

func (g *GitUtils) WithContext(ctx context.Context) *GitUtils {
	return &GitUtils{
		ctx:  ctx,
		repo: g.repo,
	}
}

func (g *GitUtils) Repository() *git.Repository {
	return g.repo
}

// MarkWorktree create a local commit with given msg and tag it with given tag.
// if detach == true, soft reset to initial head after done
func (g *GitUtils) MarkWorktree(tag string, msg string, detach bool,  matchers...GitFileMatcher) error {
	headHash, e := g.HeadCommitHash()
	if e != nil {
		return fmt.Errorf(`cannot get HEAD commit hash`)
	}

	hash, e := g.CommitIfModified(msg, matchers...)
	if e != nil {
		return e
	}

	if e := g.TagCommit(tag, hash, nil, true); e != nil {
		return e
	}

	logger.WithContext(g.ctx).Debugf(`Git: Marked current worktree as [mark_tag = %s] [commit = %v]`, tag, hash)

	if !detach {
		return nil
	}
	if e := g.ResetToCommit(headHash, false); e != nil {
		msg := fmt.Sprintf("unable to reset current branch after marking: %v. Worktree need manual clean up", e)
		logger.WithContext(g.ctx).Errorf(`Git: %s`, msg)
		return fmt.Errorf(msg)
	}

	return nil
}

// MarkCommit create a lightweight tag of given commit hash
func (g *GitUtils) MarkCommit(tag string, commitHash plumbing.Hash) error {
	if e := g.TagCommit(tag, commitHash, nil, true); e != nil {
		return e
	}
	logger.WithContext(g.ctx).Debugf(`Git: Marked current commit [%v] as lightweight tag [%s]`, commitHash, tag)
	return nil
}

// MarkedCommit returns commit hash that previously marked with given TAG
func (g *GitUtils) MarkedCommit(tag string) (plumbing.Hash, error) {
	tagRef, e := g.repo.Tag(tag)
	if e != nil {
		return plumbing.ZeroHash, fmt.Errorf(errTmplGetMarkedTag, tag, e)
	}

	t, e := g.repo.TagObject(tagRef.Hash())
	switch e {
	case nil:
		// annotated tag
		commit, e := t.Commit()
		if e != nil {
			return plumbing.ZeroHash, fmt.Errorf(errTmplGetMarkedTag, tag, e)
		}
		return commit.Hash, nil
	case plumbing.ErrObjectNotFound:
		// lightweight tag, tagRef.Hash is the commit hash
		return tagRef.Hash(), nil
	}
	return plumbing.ZeroHash, fmt.Errorf(errTmplGetMarkedTag, tag, e)
}

// TagMarkedCommit find previously marked commit using "markedTag" and re-tag it with "newTag" as name and opts
// if opts is nil, the new tag is lightweight, otherwise annotated tag
func (g *GitUtils) TagMarkedCommit(markedTag string, newTag string, opts *git.CreateTagOptions) error {
	hash, e := g.MarkedCommit(markedTag)
	if e != nil {
		return fmt.Errorf(errTmplGetMarkedTag, markedTag, e)
	}

	if e := g.TagCommit(newTag, hash, opts, true); e != nil {
		return e
	}

	tagType := "lightweight"
	if opts != nil {
		tagType = "annotated"
	}
	logger.WithContext(g.ctx).Debugf(`Git: Tagged mark [%s] as %s tag [%s]`, markedTag, tagType, newTag)
	return e
}

// ResetToMarkedCommit find previously marked commit using "markedTag" and reset current branch to the marked commit
// if current branch doesn't include the marked commit, error would return
func (g *GitUtils) ResetToMarkedCommit(markedTag string, discardChanges bool) error {
	hash, e := g.MarkedCommit(markedTag)
	if e != nil {
		return fmt.Errorf("unable to find commit with given tag %s: %v", markedTag, e)
	}

	return g.ResetToCommit(hash, discardChanges)
}

// CommitIfModified perform commit on matched files (all files if matchers not specified).
// it returns the commit hash if commit was performed or the current HEAD if commit is not performed (no files committed)
func (g *GitUtils) CommitIfModified(msg string, matchers ...GitFileMatcher) (plumbing.Hash, error) {
	worktree := g.mustWorktree()
	status, e := worktree.Status()
	if e != nil {
		return plumbing.ZeroHash, e
	}

	// filter files
	filtered := gitFilterStatus(status, matchers)
	if len(filtered) == 0 {
		return g.mustHead().Hash(), nil
	}

	// stage and commit
	for _, f := range filtered {
		if e := worktree.AddWithOptions(&git.AddOptions{Path: f}); e != nil {
			return plumbing.ZeroHash, e
		}
	}
	
	return worktree.Commit(msg, &git.CommitOptions{})
}

// TagCommit tag given commit hash.
// if the tag already exist and allowReTag == true, the existing tag would be redirected to given commit
// if opts == nil, lightweight tag is created, otherwise annotated tag
func (g *GitUtils) TagCommit(tag string, commitHash plumbing.Hash, opts *git.CreateTagOptions, allowReTag bool) error {
	commit, e := g.repo.CommitObject(commitHash)
	if e != nil {
		return e
	}

	_, e = g.repo.Tag(tag)
	switch {
	case e != nil && e != git.ErrTagNotFound && e != plumbing.ErrObjectNotFound:
		return e
	case e == nil && !allowReTag:
		return fmt.Errorf(`tag "%s" already exist and re-tagging is enabled`, tag)
	case e == nil:
		if e := g.repo.DeleteTag(tag); e != nil {
			return fmt.Errorf(`tag "%s" already exist failed to delete: %v`, tag, e)
		}
	}
	if _, e := g.repo.CreateTag(tag, commit.Hash, opts); e != nil {
		return e
	}
	return nil
}

// ResetToCommit Reset current branch to given commit hash
func (g *GitUtils) ResetToCommit(commitHash plumbing.Hash, discardChanges bool) error {
	var mode git.ResetMode
	if discardChanges {
		mode = git.HardReset
	}
	worktree := g.mustWorktree()
	return worktree.Reset(&git.ResetOptions{
		Commit: commitHash,
		Mode: mode,
	})
}

// HeadCommitHash get current Head commit
// it returns the commit hash if commit was performed or the current HEAD if commit is not performed (no files committed)
func (g *GitUtils) HeadCommitHash() (plumbing.Hash, error) {
	head, e := g.repo.Head()
	if e != nil {
		return plumbing.ZeroHash, e
	}
	return head.Hash(), nil
}

func (g *GitUtils) mustWorktree() *git.Worktree {
	worktree, e := g.repo.Worktree()
	if e != nil {
		panic(e)
	}
	return worktree
}

func (g *GitUtils) mustHead() *plumbing.Reference {
	head, e := g.repo.Head()
	if e != nil {
		panic(e)
	}
	return head
}

/**************************
	Helpers
 **************************/
func gitFilterStatus(status git.Status, matchers []GitFileMatcher) []string {
	files := make([]string, len(status))
	var i int
	for k := range status {
		files[i] = k
		i++
	}
	return gitFilterFiles(files, matchers)
}

func gitFilterFiles(files []string, matchers []GitFileMatcher) (ret []string) {
	if len(matchers) == 0 {
		return files
	}
	for _, f := range files {
		for _, m := range matchers {
			if m(f) {
				ret = append(ret, f)
				break
			}
		}
	}
	return
}


