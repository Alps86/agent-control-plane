package sqlite

const commentInsertSQL = `INSERT INTO task_comments
	(id, organization_id, task_id, content, source_kind, source_id, source_name, reference_type,
	reference_id, reference_organization_id, reference_display_name, reference_link)
	SELECT ?, t.organization_id, t.id, ?, ?, ?,
		CASE WHEN ? = 'operator' THEN 'Betreiberin'
			ELSE (SELECT a.name FROM agents a WHERE a.id = ? AND a.organization_id = t.organization_id) END,
		?, ?, ?, ?, ?
	FROM tasks t JOIN projects p ON p.id = t.project_id AND p.organization_id = t.organization_id
	JOIN organizations o ON o.id = t.organization_id
	WHERE t.organization_id = ? AND t.id = ?
	AND NOT EXISTS (SELECT 1 FROM project_archives x
		WHERE x.organization_id = t.organization_id AND x.project_id = t.project_id)
	AND ((? = 'operator' AND o.operator_id = ?) OR
		(? = 'agent' AND EXISTS (SELECT 1 FROM agents a JOIN agent_project_scopes s
			ON s.agent_id = a.id AND s.organization_id = a.organization_id
			WHERE a.id = ? AND a.organization_id = t.organization_id AND a.status = 'active'
			AND s.project_id = t.project_id AND s.can_write = 1)))
	AND (? != 'task' OR EXISTS (SELECT 1 FROM tasks r
		WHERE r.id = ? AND r.organization_id = ?))`

const commentListSQL = `SELECT COALESCE(c.id, ''), COALESCE(c.organization_id, ''),
	COALESCE(c.task_id, ''), COALESCE(c.content, ''), COALESCE(c.source_kind, ''),
	COALESCE(c.source_id, ''), COALESCE(c.source_name, ''), COALESCE(c.created_at, ''),
	COALESCE(c.updated_at, ''), COALESCE(c.reference_type, ''), COALESCE(c.reference_id, ''),
	COALESCE(c.reference_organization_id, ''), COALESCE(c.reference_display_name, ''),
	COALESCE(c.reference_link, '') FROM tasks t
	JOIN organizations o ON o.id = t.organization_id
	LEFT JOIN task_comments c ON c.organization_id = t.organization_id AND c.task_id = t.id
	WHERE t.organization_id = ? AND t.id = ?
	AND ((? = 'operator' AND o.operator_id = ?) OR
		(? = 'agent' AND EXISTS (SELECT 1 FROM agents a JOIN agent_project_scopes s
			ON s.agent_id = a.id AND s.organization_id = a.organization_id
			WHERE a.id = ? AND a.organization_id = t.organization_id AND a.status = 'active'
			AND s.project_id = t.project_id AND s.can_read = 1)))
	ORDER BY c.created_at, c.id`

const commentUpdateSQL = `UPDATE task_comments SET content = ?,
	updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
	WHERE organization_id = ? AND task_id = ? AND id = ?
	AND source_kind = ? AND source_id = ?
	AND EXISTS (SELECT 1 FROM tasks t JOIN organizations o ON o.id = t.organization_id
		WHERE t.organization_id = task_comments.organization_id AND t.id = task_comments.task_id
		AND NOT EXISTS (SELECT 1 FROM project_archives x
			WHERE x.organization_id = t.organization_id AND x.project_id = t.project_id)
		AND ((? = 'operator' AND o.operator_id = ?) OR
			(? = 'agent' AND EXISTS (SELECT 1 FROM agents a JOIN agent_project_scopes s
				ON s.agent_id = a.id AND s.organization_id = a.organization_id
				WHERE a.id = ? AND a.organization_id = t.organization_id AND a.status = 'active'
				AND s.project_id = t.project_id AND s.can_write = 1))))`
