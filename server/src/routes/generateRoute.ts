import { Router, type Router as ExpressRouter } from "express";
import { generateRules, generateSummary } from "../controller/generateController.js";

const generateRouter: ExpressRouter = Router();

generateRouter.post("/rules", generateRules);
generateRouter.post("/summary", generateSummary);

export default generateRouter;
