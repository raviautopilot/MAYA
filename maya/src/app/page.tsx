import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FullPageLoader } from '@/components/ui';

export default function Home() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    // Redirect to the correct page on mount.
    navigate(token ? '/dashboard' : '/login', { replace: true });
  }, [navigate]);

  useEffect(() => {
    // Show a loader for a split second to prevent flashing the login page
    // if the user is already authenticated.
    const timer = setTimeout(() => setLoading(false), 250);
    return () => clearTimeout(timer);
  }, []);

  // Render a loader until the redirect is complete.
  return <FullPageLoader />;
}
