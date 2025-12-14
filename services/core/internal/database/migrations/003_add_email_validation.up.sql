ALTER TABLE email_integrations ADD CONSTRAINT email_integrations_email_address_check
CHECK (email_address ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');