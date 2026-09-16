import { Routes, Route, Navigate, NavLink, Outlet, useLocation } from 'react-router-dom'
import { AppShell, Button, NoticeBanner, useTheme, type NavGroup } from 'cheval-ui'
import { CircleUser, Code, LogOut, Moon, Server, Network as NetworkIcon, HardDrive, ListTodo, SquareTerminal, Sun } from 'lucide-react'
import { clearToken, getToken } from './api'
import { Logo } from './Logo'
import Login from './pages/Login'
import VMs from './pages/VMs'
import Networks from './pages/Networks'
import Disks from './pages/Disks'
import Media from './pages/Media'
import VMDetails from './pages/VMDetails'
import Profile from './pages/Profile'
import Jobs from './pages/Jobs'
import JobDetails from './pages/JobDetails'
import Api from './pages/Api'
import Console from './pages/Console'
import { JobsProvider } from './components/JobsProvider'
import { JobsPanel } from './components/JobsPanel'
import { SessionProvider, useSession } from './hooks/useSession'

function navigation(admin: boolean): NavGroup[] {
  return [
    {
      items: [
        { to: '/', label: 'Virtual Machines', icon: Server, exact: true },
        { to: '/networks', label: 'Networks', icon: NetworkIcon },
        { to: '/disks', label: 'Disks', icon: HardDrive },
        { to: '/media', label: 'Images & ISOs', icon: HardDrive },
        { to: '/jobs', label: 'Activity', icon: ListTodo },
        { to: '/api-docs', label: 'API', icon: Code },
        ...(admin ? [{ to: '/console', label: 'Console', icon: SquareTerminal }] : []),
      ],
    },
  ]
}

function sidebarButtonClass(collapsed: boolean) {
  return `flex w-full items-center rounded-md text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/50 hover:text-sidebar-foreground ${
    collapsed ? 'h-8 justify-center' : 'gap-2.5 px-3 py-1.5 text-sm font-medium'
  }`
}

function AccountLink({ collapsed }: { collapsed: boolean }) {
  const session = useSession()
  const name = session.user?.username || (session.loading ? 'Loading…' : 'Account')

  return (
    <NavLink
      to="/profile"
      title={name}
      aria-label={name}
      className={({ isActive }) =>
        `${sidebarButtonClass(collapsed)} ${isActive ? 'bg-sidebar-accent/60 text-sidebar-foreground' : ''}`
      }
    >
      <CircleUser className="h-4 w-4 shrink-0" />
      {!collapsed && <span className="truncate">{name}</span>}
    </NavLink>
  )
}

function SidebarFooter({ collapsed }: { collapsed: boolean }) {
  const { theme, toggle } = useTheme()

  function logout() {
    clearToken()
    Object.keys(sessionStorage)
      .filter((key) => key.startsWith('maco:draft:'))
      .forEach((key) => sessionStorage.removeItem(key))
    window.location.assign('/login')
  }

  const themeLabel = theme === 'dark' ? 'Light mode' : 'Dark mode'

  return (
    <div className="space-y-0.5">
      <button onClick={toggle} title={themeLabel} aria-label={themeLabel} className={sidebarButtonClass(collapsed)}>
        {theme === 'dark' ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
        {!collapsed && themeLabel}
      </button>
      <button onClick={logout} title="Logout" aria-label="Logout" className={sidebarButtonClass(collapsed)}>
        <LogOut aria-hidden="true" className="h-3.5 w-3.5" />
        {!collapsed && 'Logout'}
      </button>
    </div>
  )
}

function SessionShell(props: Omit<React.ComponentProps<typeof AppShell>, 'nav'>) {
  const session = useSession()

  return (
    <AppShell
      {...props}
      nav={navigation(session.admin)}
      showThemeToggle={false}
      sidebarTop={({ collapsed }) => <AccountLink collapsed={collapsed} />}
      sidebarFooter={({ collapsed }) => <SidebarFooter collapsed={collapsed} />}
    >
      {session.loading ? (
        <p role="status">Loading account permissions…</p>
      ) : session.error ? (
        <NoticeBanner intent="danger">
          <div className="space-y-2">
            <p>Couldn’t load account permissions. Retry to continue.</p>
            <Button onClick={session.refresh}>Retry</Button>
          </div>
        </NoticeBanner>
      ) : (
        props.children
      )}
    </AppShell>
  )
}

function Shell() {
  const location = useLocation()
  if (!getToken()) {
    return <Navigate to={`/login?returnTo=${encodeURIComponent(location.pathname + location.search)}`} replace />
  }

  return (
    <SessionProvider>
      <JobsProvider>
        <SessionShell
          title="maco"
          logo={Logo}
          mobileNavigation={<JobsPanel />}
        >
          <Outlet />
        </SessionShell>
      </JobsProvider>
    </SessionProvider>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<Shell />}>
        <Route path="/" element={<VMs />} />
        <Route path="/networks" element={<Networks />} />
        <Route path="/disks" element={<Disks />} />
        <Route path="/media" element={<Media />} />
        <Route path="/vms/:id" element={<VMDetails />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/api-docs" element={<Api />} />
        <Route path="/console" element={<Console />} />
        <Route path="/jobs" element={<Jobs />} />
        <Route path="/jobs/:id" element={<JobDetails />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
