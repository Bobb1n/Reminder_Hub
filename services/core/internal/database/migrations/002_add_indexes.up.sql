CREATE INDEX idx_email_integrations_email_user ON email_integrations(email_address, user_id);
CREATE INDEX idx_emails_raw_processed ON emails_raw(processed);
CREATE INDEX idx_emails_raw_date_received ON emails_raw(date_received);