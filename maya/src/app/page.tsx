'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { FullPageLoader } from '@/components/ui/Loader';

export default function Home() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    // Redirect to the correct page, but only on the client-side after mount.
    router.replace(token ? '/dashboard' : '/login');
  }, [router]);

  useEffect(() => {
    // Show a loader for a split second to prevent flashing the login page
    // if the user is already authenticated.
    const timer = setTimeout(() => setLoading(false), 250);
    return () => clearTimeout(timer);
  }, []);

  // Render a loader until the redirect is complete.
  return <FullPageLoader />;
}
