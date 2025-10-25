-- Drop trigger
DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;

-- Drop indexes
DROP INDEX IF EXISTS idx_todos_user_id;
DROP INDEX IF EXISTS idx_todos_completed;
DROP INDEX IF EXISTS idx_todos_created_at;

-- Drop table
DROP TABLE IF EXISTS todos;