export const checkSession = async (): Promise<boolean> => {
  try {
    const response = await fetch('http://localhost:8080/profile', {
      method: 'GET',
      credentials: 'include'
    });
    
    return response.ok;
  } catch (error) {
    console.error('Session check failed:', error);
    return false;
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
  } else if (!hasSession && isProtectedRoute) {
    router.push('/login');
  }
};
