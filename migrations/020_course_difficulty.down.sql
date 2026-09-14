UPDATE modules SET difficulty = CASE difficulty WHEN 'beginner' THEN 'easy' WHEN 'intermediate' THEN 'medium' ELSE 'hard' END
 WHERE slug IN ('sql-easy', 'sql-medium', 'sql-hard');
UPDATE lessons SET difficulty = CASE difficulty WHEN 'beginner' THEN 'easy' WHEN 'intermediate' THEN 'medium' ELSE 'hard' END
 WHERE module_id IN (SELECT id FROM modules WHERE slug IN ('sql-easy', 'sql-medium', 'sql-hard'));
