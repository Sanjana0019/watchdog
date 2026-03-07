import Image from "next/image";

export default function WatchdogLogo({ className }: { className?: string }) {
  return (
    <Image
      src="/logo.png"
      alt="Watchdog"
      width={120}
      height={32}
      className={className}
      priority
    />
  );
}
