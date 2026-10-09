-- A lesson can state the size of the machine its lab needs.
--
-- Until now the VM's memory and CPU came from profileOf(), which reads the
-- sandbox image name and knows three sizes. An author who needed more had no
-- way to say so: the only lever was to invent an image, which also changes
-- what is installed. These two columns separate the question "what is on the
-- machine" (the image) from "how big is it" (here).
--
-- 0 means "whatever the profile gives", so every existing lesson keeps the
-- size it boots with today. The runner clamps what it is handed - FC_MAX_VMS
-- of these run on one host - so a number here is a request, not a promise.
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS vm_cpus    INT NOT NULL DEFAULT 0;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS vm_mem_mib INT NOT NULL DEFAULT 0;
