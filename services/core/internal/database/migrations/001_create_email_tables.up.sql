CREATE TABLE email_integrations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    email_address VARCHAR(255) NOT NULL,
    imap_host VARCHAR(255) NOT NULL,
    imap_port INTEGER NOT NULL,
    use_ssl BOOLEAN NOT NULL DEFAULT true,
    password TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_sync_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE emails_raw (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    from_address VARCHAR(255),
    subject TEXT,
    body_text TEXT,
    date_received TIMESTAMP WITH TIME ZONE,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_email_integrations_user_id ON email_integrations(user_id);
CREATE INDEX idx_emails_raw_user_id ON emails_raw(user_id);
CREATE INDEX idx_emails_raw_message_id ON emails_raw(message_id);