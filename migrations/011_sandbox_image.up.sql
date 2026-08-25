-- 011: point existing shell tasks at the purpose-built sandbox image.
-- Lab containers run with --network none, so tools must be baked into the
-- image (deploy/sandbox/Dockerfile) instead of installed at runtime.
UPDATE tasks SET sandbox_image = 'golearn/sandbox:latest'
 WHERE kind = 'shell' AND sandbox_image IN ('', 'ubuntu:24.04', 'base');

UPDATE lessons SET vm_image = 'golearn/sandbox:latest'
 WHERE vm_image IN ('ubuntu:24.04', 'base');
