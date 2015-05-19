// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package git defines a Git repository type that implements the repo.Repo
// interface.
package git

import (
	"errors"
	"path/filepath"
	"strings"

	g2g "github.com/libgit2/git2go"
)

// GitRepo implements the Repo interface.
type GitRepo struct {
	absPath string
	r       *g2g.Repository
	url     string
}

// New creates a new GitRepo.
func New(absPath string, url string) (*GitRepo, error) {
	// attempt opening the repository as it may already exist
	// ignore if it fails since it will be created at first call to Clone()
	r, _ := g2g.OpenRepository(absPath)

	return &GitRepo{absPath: absPath, url: url, r: r}, nil
}

// AbsPath implements the AbsPath() method of the Repo interface.
func (gr GitRepo) AbsPath() string {
	return gr.absPath
}

// URL implements the URL() method of the Repo interface.
func (gr GitRepo) URL() string {
	return gr.url
}

// Clone implements the Clone() method of the Repo interface.
func (gr GitRepo) Clone() error {
	var err error
	gr.r, err = g2g.Clone(gr.url, gr.absPath, &g2g.CloneOptions{})
	if err != nil {
		return err
	}

	return nil
}

// Update implements the Update() method of the Repo interface.
// It fetches changes from remote and performs a fast-forward on the local
// branch so as to match the remote branch.
func (gr GitRepo) Update() error {
	origin, err := gr.r.LookupRemote("origin")
	if err != nil {
		return err
	}

	err = origin.Fetch([]string{}, nil, "")
	if err != nil {
		return err
	}

	ref, err := gr.r.Head()
	if err != nil {
		return err
	}

	if !ref.IsBranch() {
		return errors.New("repository reference is not a branch (likely in a detached HEAD state)")
	}

	remoteRef, err := ref.Branch().Upstream()
	if err != nil {
		return err
	}
	_, err = ref.SetTarget(remoteRef.Target(), nil, "pull: Fast-forward")
	if err != nil {
		return err
	}

	var checkoutOpts g2g.CheckoutOpts
	checkoutOpts.Strategy = g2g.CheckoutForce

	if err = gr.r.CheckoutHead(&checkoutOpts); err != nil {
		return errors.New("failed to checkout new HEAD")
	}

	return nil
}

// hasGitExt returns true if the path ends with a ".git" extension,
// false otherwise.
func hasGitExt(path string) bool {
	return filepath.Ext(path) == ".git"
}

// cleanURL removes the trailing slashes of an URL, if any.
func cleanURL(url string) string {
	return strings.TrimSuffix(url, "/")
}
