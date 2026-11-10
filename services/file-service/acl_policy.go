package main

import (
	"context"

	"lite-nas/shared/fileaccess"
)

// authorizeRead applies an application's configured folder ACL rule to one read.
func (s *fileService) authorizeRead(ctx context.Context, subject fileaccess.Subject, path string) error {
	return s.auth.Read(ctx, subject, s.cfg.ACLPolicy.ReadDecisionPath(subject.Application, path))
}

// listingCanUseParentACL determines whether a successful directory check covers an entry.
func (s *fileService) listingCanUseParentACL(application string, directory string, entry string) bool {
	return s.cfg.ACLPolicy.ParentFolderCoversEntry(application, directory, entry)
}
