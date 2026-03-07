import type { Request, Response } from "express";
import { toolname } from "../../generated/prisma/enums.js";
import { runRuleWriterAgent } from "../agents/rule-writer.js";
import { runProjectSummariserAgent } from "../agents/project-summariser.js";
import { logDebug, logError, logInfo } from "../utils/logger.js";

const getSingleValue = (value: unknown): string | undefined => {
  if (typeof value === "string") return value;
  if (Array.isArray(value) && typeof value[0] === "string") return value[0];
  return undefined;
};

const isValidToolName = (value: unknown): value is toolname =>
  typeof value === "string" &&
  Object.values(toolname).includes(value as toolname);

export const generateRules = async (req: Request, res: Response) => {
  try {
    const { contents } = req.body;
    const selectedTool = getSingleValue(req.query.toolname);
    logDebug("generate.rules", "request", { tool: selectedTool ?? null });

    if (!isValidToolName(selectedTool)) {
      return res.status(400).send("error");
    }

    if (contents !== undefined && typeof contents !== "string") {
      return res.status(400).send("error");
    }

    const projectSummary =
      typeof contents === "string" && contents.trim().length > 0 ? contents.trim() : undefined;

    await runRuleWriterAgent(selectedTool, projectSummary);
    logInfo("generate.rules", "completed", {
      tool: selectedTool,
      hasSummary: Boolean(projectSummary),
    });

    return res.status(200).send("success");
  } catch (error) {
    logError("generate.rules", "failed", { error: String(error) });
    return res.status(500).send("error");
  }
};

export const generateSummary = async (req: Request, res: Response) => {
  try {
    const rootFromQuery = getSingleValue(req.query.path);
    const rootFromBodyPath =
      typeof req.body?.path === "string" ? req.body.path : undefined;
    const rootFromBodyProjectRoot =
      typeof req.body?.projectRoot === "string" ? req.body.projectRoot : undefined;

    const projectRoot = rootFromQuery ?? rootFromBodyProjectRoot ?? rootFromBodyPath;
    if (!projectRoot || projectRoot.trim().length === 0) {
      return res.status(400).send("error");
    }

    logDebug("generate.summary", "request", { projectRoot: projectRoot.trim() });
    await runProjectSummariserAgent(projectRoot.trim());
    logInfo("generate.summary", "completed", { projectRoot: projectRoot.trim() });
    return res.status(200).send("success");
  } catch (error) {
    logError("generate.summary", "failed", { error: String(error) });
    return res.status(500).send("error");
  }
};
