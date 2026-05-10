CREATE TYPE job_status AS ENUM (
    'pending',
    'processing',
    'succeeded',
    'failed',
    'dead'
);

CREATE TABLE retry_jobs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id       VARCHAR(255) NOT NULL,
    amount           DECIMAL(10, 2) NOT NULL,
    currency         VARCHAR(3) NOT NULL DEFAULT 'INR',
    status           job_status NOT NULL DEFAULT 'pending',
    retry_count      INT NOT NULL DEFAULT 0,
    max_retries      INT NOT NULL DEFAULT 3,
    next_retry_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_error       TEXT,
    webhook_url      VARCHAR(500),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_retry_job_status ON retry_jobs(status);
CREATE INDEX idx_retry_job_next_retry_at ON retry_jobs(next_retry_at);
CREATE INDEX idx_retry_job_status_next_retry ON retry_jobs(status, next_retry_at);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_retry_jobs_updated_at
    BEFORE UPDATE ON retry_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();