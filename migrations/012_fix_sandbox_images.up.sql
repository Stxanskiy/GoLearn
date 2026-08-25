-- 012: normalise sandbox images.
-- The imported courses carried platform-side labels ("docker", "kuber",
-- "golearn/linux:latest") in tasks.sandbox_image; `docker run` cannot start
-- those, so every such lab failed to open. Point them at the image we ship.
UPDATE tasks SET sandbox_image = 'golearn/sandbox:latest'
 WHERE kind = 'shell'
   AND sandbox_image NOT IN ('golearn/sandbox:latest', 'golearn/sandbox-pg:latest');

UPDATE lessons SET vm_image = 'golearn/sandbox:latest'
 WHERE vm_image <> '' AND vm_image NOT IN ('golearn/sandbox:latest', 'golearn/sandbox-pg:latest');
