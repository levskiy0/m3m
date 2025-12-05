import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

// Redirect to dashboard with logs tab
export function LogsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();

  useEffect(() => {
    if (projectId) {
      navigate(`/projects/${projectId}`, {
        replace: true,
        state: { tab: 'logs' }
      });
    }
  }, [projectId, navigate]);

  return null;
}
