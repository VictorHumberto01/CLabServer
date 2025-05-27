-- Drop indexes
DROP INDEX IF EXISTS idx_submissions_user_id;
DROP INDEX IF EXISTS idx_submissions_task_id;
DROP INDEX IF EXISTS idx_test_cases_task_id;
DROP INDEX IF EXISTS idx_tasks_room_id;
DROP INDEX IF EXISTS idx_rooms_teacher_id;

-- Drop tables
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS room_students;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users; 