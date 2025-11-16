CREATE TABLE IF NOT EXISTS pr_requests (
    pull_request_id TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(user_id),
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT max_reviewers CHECK (array_length(assigned_reviewers, 1) <= 2)
);

CREATE INDEX IF NOT EXISTS idx_pr_requests_author_id ON pr_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_requests_status ON pr_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_requests_reviewers ON pr_requests USING GIN(assigned_reviewers);
