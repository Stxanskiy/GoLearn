-- 020: courses and lessons use the beginner..expert scale; easy/medium/hard is for tasks.
UPDATE modules SET difficulty = CASE difficulty WHEN 'easy' THEN 'beginner' WHEN 'medium' THEN 'intermediate' ELSE 'advanced' END
 WHERE difficulty IN ('easy', 'medium', 'hard');
UPDATE lessons SET difficulty = CASE difficulty WHEN 'easy' THEN 'beginner' WHEN 'medium' THEN 'intermediate' ELSE 'advanced' END
 WHERE difficulty IN ('easy', 'medium', 'hard');
