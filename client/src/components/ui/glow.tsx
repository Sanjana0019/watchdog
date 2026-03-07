import { cn } from "@/lib/utils";

export default function Glow({
  className,
  variant = "top",
}: {
  className?: string;
  variant?: "top" | "center";
}) {
  return (
    <div
      className={cn(
        "pointer-events-none absolute inset-x-0 -z-10 mx-auto h-44 w-[80%] rounded-full bg-gradient-to-r from-cyan-300/35 via-sky-400/35 to-teal-300/35 blur-3xl",
        variant === "top" ? "top-0" : "top-1/3",
        className,
      )}
    />
  );
}
