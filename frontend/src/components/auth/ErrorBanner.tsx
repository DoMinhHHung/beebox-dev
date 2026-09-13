type ErrorBannerProps = { title: string; message: string };

export function ErrorBanner({ title, message }: ErrorBannerProps) {
  return (
    <div className="w-full rounded-lg bg-error-container/60 border border-error/20 px-4 py-3 text-left" role="alert">
      <p className="text-sm font-medium text-on-error-container">{title}</p>
      <p className="text-xs text-on-error-container/80 mt-0.5">{message}</p>
    </div>
  );
}
