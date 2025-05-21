import { NavLink, useLocation, useNavigate } from 'react-router-dom';
import { 
  Shield, 
  LayoutDashboard, 
  Bell, 
  FileText, 
  Settings, 
  LogOut,
  X
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';

interface SidebarProps {
  isOpen: boolean;
  isMobile: boolean;
  closeSidebar: () => void;
}

export default function Sidebar({ isOpen, isMobile, closeSidebar }: SidebarProps) {
  const { logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    {
      name: 'Dashboard',
      path: '/dashboard',
      icon: <LayoutDashboard className="h-5 w-5" />,
    },
    {
      name: 'Alerts & Logs',
      path: '/alerts',
      icon: <Bell className="h-5 w-5" />,
    },
    {
      name: 'Policies',
      path: '/policies',
      icon: <FileText className="h-5 w-5" />,
    },
    {
      name: 'Settings',
      path: '/settings',
      icon: <Settings className="h-5 w-5" />,
    },
  ];

  // If mobile and sidebar is closed, return null
  if (isMobile && !isOpen) {
    return null;
  }

  return (
    <>
      {/* Mobile overlay */}
      {isMobile && isOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/50" 
          onClick={closeSidebar}
        />
      )}

      {/* Sidebar */}
      <div 
        className={cn(
          "z-50 flex h-full flex-col border-r bg-card transition-all duration-300",
          isOpen ? "w-64" : "w-0",
          isMobile && isOpen ? "fixed left-0" : "",
          isMobile && !isOpen ? "hidden" : ""
        )}
      >
        {/* Logo and close button */}
        <div className="flex h-16 items-center justify-between border-b px-4">
          <div className="flex items-center space-x-2">
            <Shield className="h-6 w-6 text-primary" />
            <span className="text-lg font-bold">SecureFlow</span>
          </div>
          {isMobile && (
            <Button 
              size="icon" 
              variant="ghost" 
              onClick={closeSidebar}
            >
              <X className="h-5 w-5" />
            </Button>
          )}
        </div>

        {/* Navigation items */}
        <ScrollArea className="flex-1 py-4">
          <nav className="space-y-1 px-2">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                onClick={isMobile ? closeSidebar : undefined}
                className={({ isActive }) =>
                  cn(
                    "flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )
                }
              >
                {item.icon}
                <span>{item.name}</span>
              </NavLink>
            ))}
          </nav>
        </ScrollArea>

        {/* Footer */}
        <div className="border-t p-4">
          <Button 
            variant="outline" 
            className="w-full justify-start space-x-2" 
            onClick={handleLogout}
          >
            <LogOut className="h-4 w-4" />
            <span>Logout</span>
          </Button>
        </div>
      </div>
    </>
  );
}