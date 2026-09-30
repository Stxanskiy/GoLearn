-- Course and section icons are gone from the product: they were a small square
-- badge floating half over the cover and half over the card, repeating what the
-- cover already showed, and an upload slot in the editor that had to be filled
-- for every course. The cover carries the identity now.
--
-- The emoji fallback went with them — it was only ever shown until an icon was
-- uploaded. Simulators keep their own emoji: that is part of the scenario, not
-- an uploaded image.
ALTER TABLE modules         DROP COLUMN IF EXISTS icon_url;
ALTER TABLE specializations DROP COLUMN IF EXISTS icon;
ALTER TABLE specializations DROP COLUMN IF EXISTS icon_url;
