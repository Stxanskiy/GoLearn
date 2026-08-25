UPDATE tasks   SET sandbox_image = 'ubuntu:24.04' WHERE sandbox_image = 'golearn/sandbox:latest';
UPDATE lessons SET vm_image      = 'ubuntu:24.04' WHERE vm_image      = 'golearn/sandbox:latest';
