type LogLevel = "debug" | "info" | "warn" | "error";

const debugEnabled =
  process.env.DEBUG_AI === "1" ||
  process.env.DEBUG_AI === "true" ||
  process.env.NODE_ENV !== "production";

const formatMeta = (meta?: Record<string, unknown>): string => {
  if (!meta) return "";
  try {
    return ` ${JSON.stringify(meta)}`;
  } catch {
    return " [meta-unserializable]";
  }
};

const emit = (level: LogLevel, scope: string, message: string, meta?: Record<string, unknown>) => {
  if (level === "debug" && !debugEnabled) return;

  const line = `[${new Date().toISOString()}] [${level}] [${scope}] ${message}${formatMeta(meta)}`;
  if (level === "error") {
    console.error(line);
    return;
  }
  if (level === "warn") {
    console.warn(line);
    return;
  }
  console.log(line);
};

export const logDebug = (scope: string, message: string, meta?: Record<string, unknown>) =>
  emit("debug", scope, message, meta);

export const logInfo = (scope: string, message: string, meta?: Record<string, unknown>) =>
  emit("info", scope, message, meta);

export const logWarn = (scope: string, message: string, meta?: Record<string, unknown>) =>
  emit("warn", scope, message, meta);

export const logError = (scope: string, message: string, meta?: Record<string, unknown>) =>
  emit("error", scope, message, meta);

