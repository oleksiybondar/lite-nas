package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	authcontract "lite-nas/shared/contracts/auth"
)

const captionJobLifetime = 7 * 24 * time.Hour

// captionGrantCleanupDeadline covers a grant issued at the end of a job's lifetime.
func captionGrantCleanupDeadline(created time.Time) time.Time {
	return created.Add(2 * captionJobLifetime)
}

// issueCaptionGrant asks Auth Service for a grant limited to this job's paths.
func (m *scanManager) issueCaptionGrant(ctx context.Context, job scanJob, token string, paths []string) (string, error) {
	if m.grants == nil || len(paths) == 0 || len(paths) > 10000 {
		return "", errors.New("caption grant service unavailable")
	}
	hashes := make([]string, 0, len(paths))
	for _, path := range paths {
		digest := sha256.Sum256([]byte(path))
		hashes = append(hashes, hex.EncodeToString(digest[:]))
	}
	var response authcontract.CaptionGrantIssueResponse
	if err := m.grants.Request(ctx, authcontract.CaptionGrantIssueRPCSubject,
		authcontract.CaptionGrantIssueRequest{UserAccessToken: token, JobID: job.ID, PathHashes: hashes}, &response); err != nil {
		return "", err
	}
	if len(response.Grant) != 43 || !response.ExpiresAt.After(time.Now()) {
		return "", errors.New("caption grant denied")
	}
	return response.Grant, nil
}

// saveCaptionGrant persists the credential and exact source list for restart.
func (m *scanManager) saveCaptionGrant(ctx context.Context, job scanJob, grant string, paths []string) error {
	tx, err := m.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `INSERT INTO caption_scan_paths(job_id,position,path) VALUES(?,?,?)`)
	if err != nil {
		return err
	}
	for index, path := range paths {
		if _, err := statement.ExecContext(ctx, job.ID, index, path); err != nil {
			_ = statement.Close()
			return err
		}
	}
	if err := statement.Close(); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE media_scan_jobs SET grant_token=?,resume_pending=1,status='RUNNING',total_items=? WHERE id=?`, grant, len(paths), job.ID); err != nil {
		return err
	}
	return tx.Commit()
}

// captionJobPaths reloads the fixed ordered source list after a service restart.
func (m *scanManager) captionJobPaths(ctx context.Context, jobID string) ([]string, error) {
	rows, err := m.store.db.QueryContext(ctx, `SELECT path FROM caption_scan_paths WHERE job_id=? ORDER BY position`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paths := []string{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

// captionJobGrant reads a credential only inside the trusted scan worker.
func (m *scanManager) captionJobGrant(ctx context.Context, jobID string) (string, error) {
	var grant sql.NullString
	err := m.store.db.QueryRowContext(ctx, `SELECT grant_token FROM media_scan_jobs WHERE id=?`, jobID).Scan(&grant)
	if err != nil || !grant.Valid || len(grant.String) != 43 {
		return "", errors.New("caption grant unavailable")
	}
	return grant.String, nil
}

// revokeCaptionGrant removes the authorization when a job ends.
func (m *scanManager) revokeCaptionGrant(ctx context.Context, grant string) error {
	if m.grants == nil || grant == "" {
		return nil
	}
	var response authcontract.CaptionGrantRevokeResponse
	return m.grants.Request(ctx, authcontract.CaptionGrantRevokeRPCSubject,
		authcontract.CaptionGrantRevokeRequest{Grant: grant}, &response)
}

// cleanCaptionGrant retries revocation until Auth responds or the grant expires.
func (m *scanManager) cleanCaptionGrant(jobID, grant string, deadline time.Time) {
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := m.revokeCaptionGrant(ctx, grant)
		cancel()
		if err == nil {
			_, err = m.store.db.Exec(`UPDATE media_scan_jobs SET grant_token=NULL WHERE id=? AND grant_token=?`, jobID, grant)
			if err == nil {
				return
			}
		}
		_, _ = fmt.Fprintf(os.Stderr, "moments: caption grant cleanup for job %s will retry: %v\n", jobID, err)
		time.Sleep(time.Minute)
	}
	// Auth rejects the grant after its seven-day lifetime, even if it was never revoked.
	if _, err := m.store.db.Exec(`UPDATE media_scan_jobs SET grant_token=NULL WHERE id=? AND grant_token=?`, jobID, grant); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moments: could not clear expired caption grant for job %s: %v\n", jobID, err)
	}
}
