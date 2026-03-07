import { createOllama } from "ai-sdk-ollama";
import { aisdk } from '@openai/agents-extensions';

const ollamaClient = createOllama({
    baseURL: process.env.OLLAMA_BASE_URL || 'http://localhost:11434',
    apiKey: 'ollama'
});

const model = aisdk(ollamaClient("glm-5:cloud"));

export default model;