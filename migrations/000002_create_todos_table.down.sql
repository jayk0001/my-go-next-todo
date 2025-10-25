CREATE TABLE IF NOT EXISTS todos (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) on DELETE CASCADE,
  title VARCHAR(500) NOT NULL,
  description TEXT,
  completed BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
);

-- Create Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_todos_user_id on todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_completed on todos(completed);
CREATE INDEX IF NOT EXISTS idx_todos_created_at on todos(created_at);

-- Create trigger for todos table
CREATE TRIGEER update_todos_updated_at
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION update_updatd_at_column();


