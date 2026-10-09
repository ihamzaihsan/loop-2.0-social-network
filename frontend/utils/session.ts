import { API } from "./api";
export const checkSession = async (): Promise<boolean | null> => {
  try {
    const response = await fetch(`${API}/profile`, {
      method: 'GET',
      credentials: 'include'
    });
    
    if (response.ok) return true;
    if (response.status === 401) return false;
    // Rate limits and temporary server failures do not mean the session expired.
    return null;
  } catch (error) {
    console.error('Session check failed:', error);
    return null;
  }
};

// Redirect based on session status
export const redirectBasedOnSession = async (
  router: any, 
  isProtectedRoute: boolean = false
) => {
  const hasSession = await checkSession();
  
  if (hasSession && !isProtectedRoute) {
    router.push('/home');
  } else if (hasSession === false && isProtectedRoute) {
    router.push('/login');
  }
};
