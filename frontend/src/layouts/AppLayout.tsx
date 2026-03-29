import { useState } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  Server,
  Database,
  BarChart3,
  BookOpen,
  Menu,
  Network,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';

interface NavItem {
  to: string;
  icon: typeof LayoutDashboard;
  label: string;
  end?: boolean;
}

const navItems: NavItem[] = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard', end: true },
  { to: '/design', icon: Server, label: 'Design' },
  { to: '/catalog', icon: Database, label: 'Catalog' },
  { to: '/metrics', icon: BarChart3, label: 'Metrics' },
  { to: '/knowledge', icon: BookOpen, label: 'Knowledge' },
];

function NavItems({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <nav className="flex flex-1 flex-col gap-0.5 px-2">
      {navItems.map(({ to, icon: Icon, label, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              'group flex h-8 items-center gap-2.5 rounded-md px-2.5 text-sm font-medium transition-colors',
              isActive
                ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
            )
          }
        >
          <Icon className="size-4 shrink-0" />
          <span>{label}</span>
        </NavLink>
      ))}
    </nav>
  );
}

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <div className="flex h-full flex-col">
      {/* Logo */}
      <div className="flex h-12 items-center gap-2.5 border-b border-sidebar-border px-4">
        <div className="flex size-6 items-center justify-center rounded-md bg-sidebar-primary">
          <Network className="size-3.5 text-sidebar-primary-foreground" />
        </div>
        <span className="font-mono text-sm font-semibold tracking-tight text-sidebar-foreground">
          fabrik
        </span>
        <span className="ml-auto rounded px-1 py-0.5 text-[10px] font-medium text-muted-foreground ring-1 ring-border">
          beta
        </span>
      </div>

      {/* Nav */}
      <div className="flex flex-1 flex-col gap-1 py-3">
        <div className="px-4 pb-1">
          <span className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
            Navigation
          </span>
        </div>
        <NavItems onNavigate={onNavigate} />
      </div>

      {/* Footer */}
      <div className="border-t border-sidebar-border px-4 py-3">
        <p className="text-[11px] text-muted-foreground/50">
          Datacenter topology designer
        </p>
      </div>
    </div>
  );
}

export default function AppLayout() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Desktop sidebar */}
      <aside className="hidden w-[220px] shrink-0 flex-col border-r border-sidebar-border bg-sidebar lg:flex">
        <SidebarContent />
      </aside>

      {/* Mobile sidebar */}
      <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
        <SheetTrigger
          render={
            <Button
              variant="ghost"
              size="icon"
              className="fixed left-3 top-3 z-40 lg:hidden"
            />
          }
        >
          <Menu className="size-4" />
          <span className="sr-only">Toggle menu</span>
        </SheetTrigger>
        <SheetContent side="left" className="w-[220px] p-0">
          <SidebarContent onNavigate={() => setMobileOpen(false)} />
        </SheetContent>
      </Sheet>

      {/* Main content */}
      <main className="flex flex-1 flex-col overflow-hidden">
        <div className="flex h-full flex-col overflow-y-auto">
          {/* Mobile header spacer */}
          <div className="flex h-12 shrink-0 items-center border-b border-border px-4 lg:hidden">
            <div className="ml-10 flex items-center gap-2">
              <Network className="size-4 text-muted-foreground" />
              <span className="font-mono text-sm font-semibold">fabrik</span>
            </div>
          </div>
          <div className="flex-1 p-6">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
