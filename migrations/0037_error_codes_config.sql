-- +goose Up
CREATE TABLE IF NOT EXISTS error_codes (
    code VARCHAR(50) PRIMARY KEY,
    description TEXT NOT NULL,
    http_status INT NOT NULL,
    category VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO error_codes (code, description, http_status, category) VALUES
    ('ERR_UNAUTHORIZED', 'Unauthorized access', 401, 'auth'),
    ('ERR_FORBIDDEN', 'Forbidden resource', 403, 'auth'),
    ('ERR_NOT_FOUND', 'Resource not found', 404, 'client'),
    ('ERR_INVALID_INPUT', 'Invalid input data', 400, 'client'),
    ('ERR_CONFLICT', 'Resource conflict', 409, 'client'),
    ('ERR_RATE_LIMITED', 'Rate limit exceeded', 429, 'rate_limit'),
    ('ERR_INTERNAL_ERROR', 'Internal server error', 500, 'server'),
    ('ERR_DATABASE_ERROR', 'Database operation failed', 500, 'server'),
    ('ERR_EXTERNAL_SERVICE', 'External service error', 503, 'server'),
    ('ERR_VALIDATION_FAILED', 'Validation failed', 400, 'client'),
    ('ERR_SESSION_EXPIRED', 'Session expired', 401, 'auth'),
    ('ERR_IP_NOT_WHITELISTED', 'IP not whitelisted', 403, 'auth')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS error_codes;
