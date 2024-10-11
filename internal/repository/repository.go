package repository

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/kam9lo/gover/internal/version"
)

// Repository works on git repository and handles all features related strictly
// with git.
type Repository struct {
	git *git.Repository
}

// Open returns git repository if exists, otherwise fails with an
// error.
func Open(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}
	return &Repository{git: repo}, nil
}

// LatestTag returns latest known in repository tag.
func (r *Repository) LatestTag() (string, error) {
	tags, err := r.latestTags()
	if err != nil {
		return "", fmt.Errorf("latest tag: %w", err)
	}

	var latestTag *version.Version
	for _, tag := range tags {
		v, err := version.New(tag.Name().Short())
		if err != nil {
			continue
		}

		if latestTag == nil || v.IsGreater(latestTag) {
			latestTag = v
		}
	}

	if latestTag == nil {
		return tags[0].Name().Short(), nil
	}
	return latestTag.String(), nil
}

// CreateTag is an equivalent to "git tag <name>".
func (r *Repository) CreateTag(name string) error {
	log, err := r.git.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}

	var latestCommit *object.Commit
	if err = log.ForEach(func(c *object.Commit) error {
		latestCommit = c
		return io.EOF
	}); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	if latestCommit == nil {
		return fmt.Errorf("latest commit not found")
	}

	_, err = r.git.CreateTag(name, latestCommit.Hash, nil)
	return err
}

// FeatureCommits returns commits since last known tag.
func (r *Repository) FeatureCommits() ([]string, error) {
	commits, err := r.featureCommits()
	if err != nil {
		return nil, err
	}

	messages := make([]string, 0, len(commits))
	for _, c := range commits {
		messages = append(messages, strings.Trim(c.Message, "\n"))
	}

	return messages, nil
}

func (r *Repository) featureCommits() ([]*object.Commit, error) {
	log, err := r.git.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	tag, err := r.latestTags()
	if err != nil {
		return nil, fmt.Errorf("latest tag: %w", err)
	}

	taggedCommit, err := r.taggedCommit(tag[0].Hash())
	if err != nil {
		return nil, fmt.Errorf("tagged commit: %w", err)
	}

	commits := []*object.Commit{}
	for {
		c, err := log.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		if c.Hash == taggedCommit.Hash {
			return commits, nil
		}

		commits = append(commits, c)
	}
	return nil, ErrCommitNotFound
}

func (r *Repository) latestTags() ([]*plumbing.Reference, error) {
	tags, err := r.git.Tags()
	if err != nil {
		return nil, fmt.Errorf("git tags: %w", err)
	}
	defer tags.Close()

	taggedCommits := map[plumbing.Hash][]*plumbing.Reference{}
	if err = tags.ForEach(func(ref *plumbing.Reference) error {
		// Both annotated and unannotated tags are supported.
		commit, err := r.taggedCommit(ref.Hash())
		if err != nil {
			return nil
		}

		taggedCommits[commit.Hash] = append(taggedCommits[commit.Hash], ref)

		return nil
	}); err != nil {
		return nil, err
	}

	commits, err := r.git.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	defer commits.Close()

	var result []*plumbing.Reference
	if err := commits.ForEach(func(c *object.Commit) error {
		if tags, found := taggedCommits[c.Hash]; found {
			result = tags
			return io.EOF
		}
		return nil
	}); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("for each in git log: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return result, nil
}

func (r *Repository) headCommit() (*object.Commit, error) {
	headRef, err := r.git.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}
	return r.git.CommitObject(headRef.Hash())
}

func (r *Repository) taggedCommit(hash plumbing.Hash) (commit *object.Commit, err error) {
	// Unannotated tag commit.
	commit, err = object.GetCommit(r.git.Storer, hash)
	if err == nil {
		return
	}
	// Annotated tag commit.
	t, err := r.git.TagObject(hash)
	if err != nil {
		return nil, fmt.Errorf("annotated tag not found: %w", err)
	}

	commit, err = r.git.CommitObject(t.Target)
	if err != nil {
		return nil, fmt.Errorf("annotated tag commit not found: %w", err)
	}

	return
}

// ErrCommitNotFound indicates missing commits since last tag to generate new
// tag.
var ErrCommitNotFound = errors.New("commit not found")
