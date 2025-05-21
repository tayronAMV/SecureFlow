import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { Home } from 'lucide-react';

export default function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <div className="flex min-h-screen flex-col items-center justify-center text-center">
      <div className="space-y-4 px-4">
        <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl">404 - Page Not Found</h1>
        <p className="mx-auto max-w-[600px] text-muted-foreground md:text-xl/relaxed">
          We couldn't find the page you were looking for.
        </p>
        <Button onClick={() => navigate('/')} className="mt-4">
          <Home className="mr-2 h-4 w-4" /> Go to Dashboard
        </Button>
      </div>
    </div>
  );
}