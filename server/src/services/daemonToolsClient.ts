import { toolname } from "../../generated/prisma/enums.js";

type ReadToolResponse = {
  filepath: string;
  contents: string;
};

type DirenumToolResponse = {
  contents: string;
};

const daemonBaseUrl = process.env.DAEMON_BASE_URL ?? "http://localhost:4000";

const buildUrl = (pathname: string, query: Record<string, string>): string => {
  const url = new URL(pathname, daemonBaseUrl);
  for (const [key, value] of Object.entries(query)) {
    url.searchParams.set(key, value);
  }
  return url.toString();
};

const parseSuccessOrError = async (response: Response): Promise<"success" | "error"> => {
  const raw = (await response.text()).trim().toLowerCase();
  if (raw === "success") return "success";
  return "error";
};

export const readWithToolApi = async (path: string): Promise<ReadToolResponse> => {
  const response = await fetch(buildUrl("/tools/read", { path }), {
    method: "GET",
  });

  if (!response.ok) {
    throw new Error(`Daemon read failed with status ${response.status}`);
  }

  const payload = (await response.json()) as Partial<ReadToolResponse>;
  if (typeof payload.filepath !== "string" || typeof payload.contents !== "string") {
    throw new Error("Daemon read response is invalid");
  }

  return {
    filepath: payload.filepath,
    contents: payload.contents,
  };
};

export const writeWithToolApi = async (path: string, contents: string): Promise<"success" | "error"> => {
  const response = await fetch(buildUrl("/tools/write", { path }), {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({ contents }),
  });

  if (!response.ok) return "error";
  return parseSuccessOrError(response);
};

export const editWithToolApi = async (
  oldContents: string,
  newContents: string,
): Promise<"success" | "error"> => {
  const response = await fetch(buildUrl("/tools/edit", {}), {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({ oldContents, newContents }),
  });

  if (!response.ok) return "error";
  return parseSuccessOrError(response);
};

export const restartToolWithApi = async (tool: toolname): Promise<"success" | "error"> => {
  const response = await fetch(buildUrl("/tools/restart", { toolname: tool }), {
    method: "GET",
  });

  if (!response.ok) return "error";
  return parseSuccessOrError(response);
};

export const direnumWithToolApi = async ( level: number, path: string): Promise<DirenumToolResponse> => {
  const response = await fetch(buildUrl("/tools/direnum", { level: level.toString(), path }), {
    method: "GET",
  });

  if (!response.ok) {
    throw new Error(`Daemon direnum failed with status ${response.status}`);
  }

  const payload = (await response.json()) as Partial<DirenumToolResponse>;
  if (typeof payload.contents !== "string") {
    throw new Error("Daemon direnum response is invalid");
  }

  return {
    contents: payload.contents,
  };
};
