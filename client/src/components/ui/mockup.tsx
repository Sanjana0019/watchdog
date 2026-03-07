import { cn } from "@/lib/utils";

export function MockupFrame({
  children,
  className,
  size = "small",
}: {
  children: React.ReactNode;
  className?: string;
  size?: "small" | "base";
}) {
  return (
    <div
      className={cn(
        "mx-auto w-full overflow-hidden rounded-2xl border border-zinc-200 bg-white/70 p-2 shadow-2xl shadow-zinc-900/10 dark:border-zinc-800 dark:bg-zinc-950/70",
        size === "small" ? "max-w-5xl" : "max-w-6xl",
        className,
      )}
    >
      {children}
    </div>
  );
}

export function Mockup({
  children,
  className,
  type,
}: {
  children: React.ReactNode;
  className?: string;
  type?: "responsive";
}) {
  return (
    <div
      className={cn("overflow-hidden rounded-xl", className)}
      data-type={type}
    >
      {children}
    </div>
  );
}
