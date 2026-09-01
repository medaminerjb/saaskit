import { Outlet, Link, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import Logo from './ui/logo';

export default function Layout() {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  if (!isAuthenticated) {
    return null;
  }

  const getNavLinkClass = ({ isActive }: { isActive: boolean }) =>
    `inline-flex items-center px-1 pt-1 border-b-2 text-sm transition-colors duration-200 ${
      isActive
        ? 'border-primary-600 text-primary-600 font-semibold'
        : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 font-medium'
    }`;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200 shadow-xs sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <Link to="/" className="flex items-center space-x-2">
                  <div className="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
                    <Logo />
                  </div>
                  <span className="text-xl font-bold text-gray-900 tracking-tight">SaaSKit Admin</span>
                </Link>
              </div>
              <div className="hidden sm:ml-8 sm:flex sm:space-x-8">
                <NavLink to="/" end className={getNavLinkClass}>
                  Dashboard
                </NavLink>
                <NavLink to="/users" className={getNavLinkClass}>
                  Users
                </NavLink>
                <NavLink to="/tenants" className={getNavLinkClass}>
                  Tenants
                </NavLink>
                <NavLink to="/api-keys" className={getNavLinkClass}>
                  API Keys
                </NavLink>
                <NavLink to="/audit" className={getNavLinkClass}>
                  Audit Log
                </NavLink>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-3">
                <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center">
                  <span className="text-sm font-medium text-primary-600">
                    {user?.email?.charAt(0).toUpperCase()}
                  </span>
                </div>
                <span className="text-sm text-gray-700">{user?.email}</span>
              </div>
              <button
                onClick={handleLogout}
                className="text-sm text-gray-500 hover:text-gray-700 font-medium transition-colors duration-200"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-8 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}
