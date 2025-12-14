CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    email_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority VARCHAR(20) DEFAULT 'medium',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_email_id ON tasks(email_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_deadline ON tasks(deadline);
CREATE INDEX idx_tasks_user_status ON tasks(user_id, status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_tasks_priority ON tasks(priority);

CREATE UNIQUE INDEX idx_tasks_email_user_unique ON tasks(email_id, user_id);