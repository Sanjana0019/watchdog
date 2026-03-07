import Image from "next/image";

export default function Screenshot({
  srcLight,
  alt,
  width,
  height,
  className,
}: {
  srcLight: string;
  srcDark?: string;
  alt: string;
  width: number;
  height: number;
  className?: string;
}) {
  return (
    <Image
      src={srcLight}
      alt={alt}
      width={width}
      height={height}
      className={className}
      priority
    />
  );
}
