import Link from "next/link";

type AuthShellProps = { children: React.ReactNode; showFooter?: boolean };

export function AuthShell({ children, showFooter = true }: AuthShellProps) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-between bg-background px-6 py-8">
      <header className="w-full flex items-center justify-center pt-2">
        <Link href="/sign-in" className="font-semibold text-lg tracking-tight text-on-surface hover:opacity-80 transition-opacity">
          BeeBox
        </Link>
      </header>
      <main className="w-full flex-1 flex items-center justify-center py-10">{children}</main>
      {showFooter && (
        <footer className="w-full pb-2 flex flex-col sm:flex-row items-center justify-center gap-2 text-on-surface-variant text-xs">
          <div className="flex items-center gap-1.5">
            <svg className="w-3.5 h-3.5 text-primary" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
              <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
            </svg>
            <span>Secured by <span className="font-medium text-on-surface">BeeBox Identity</span></span>
          </div>
          <div className="flex items-center gap-3 mt-1 sm:mt-0 sm:ml-4">
            <a href="#" className="hover:text-on-surface transition-colors">Privacy</a>
            <span className="text-outline-variant">•</span>
            <a href="#" className="hover:text-on-surface transition-colors">Terms</a>
            <span className="text-outline-variant">•</span>
            <a href="#" className="hover:text-on-surface transition-colors">Support</a>
          </div>
        </footer>
      )}
    </div>
  );
}
