-- 004: Interactive shell tasks (Linux/DevOps sandbox).

-- Task kind: 'go' (compiled Go runner) | 'shell' (sandbox terminal).
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'go';

-- Container image used for a shell task (empty -> server default, e.g. ubuntu:24.04).
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS sandbox_image VARCHAR(120) NOT NULL DEFAULT '';

-- Setup script run once when the sandbox session is created (prepares files).
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS setup_script TEXT NOT NULL DEFAULT '';

-- Check script: exit code 0 means the task is solved.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS check_script TEXT NOT NULL DEFAULT '';
