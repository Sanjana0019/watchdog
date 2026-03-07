import { cn } from "@/lib/utils";

export function Section({
  className,
  children,
}: {
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section className={cn("px-4 py-12 sm:px-6", className)}>
      {children}
    </section>
  );
}
