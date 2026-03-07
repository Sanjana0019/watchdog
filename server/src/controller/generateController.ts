import type { Request, Response } from "express";
import { toolname } from "../../generated/prisma/enums.js";
import { runRuleWriterAgent } from "../agents/rule-writer.js";
import { runProjectSummariserAgent } from "../agents/project-summariser.js";

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

    if (!isValidToolName(selectedTool)) {
      return res.status(400).send("error");
    }

    if (typeof contents !== "string" || contents.trim().length === 0) {
      return res.status(400).send("error");
    }

    await runRuleWriterAgent(selectedTool, contents);

    return res.status(200).send("success");
  } catch {
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

    await runProjectSummariserAgent(projectRoot.trim());
    return res.status(200).send("success");
  } catch {
    return res.status(500).send("error");
  }
};
