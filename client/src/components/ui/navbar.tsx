import { cn } from "@/lib/utils";

export function Navbar({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <nav
      className={cn("flex items-center justify-between gap-6 py-4", className)}
    >
      {children}
    </nav>
  );
}

export function NavbarLeft({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex items-center gap-6", className)}>{children}</div>
  );
}

export function NavbarRight({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex items-center gap-4", className)}>{children}</div>
  );
}
