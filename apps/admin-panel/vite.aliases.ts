import path from "node:path";

export const createViteAliases = (rootDir: string): Record<string, string> => {
  return {
    "@app": path.resolve(rootDir, "src/app"),
    "@assets": path.resolve(rootDir, "src/assets"),
    "@components": path.resolve(rootDir, "src/components"),
    "@configs": path.resolve(rootDir, "src/configs"),
    "@contexts": path.resolve(rootDir, "src/contexts"),
    "@domain": path.resolve(rootDir, "src/domain"),
    "@helpers": path.resolve(rootDir, "src/helpers"),
    "@hooks": path.resolve(rootDir, "src/hooks"),
    "@models": path.resolve(rootDir, "src/models"),
    "@pages": path.resolve(rootDir, "src/pages"),
    "@providers": path.resolve(rootDir, "src/providers"),
    "@routes": path.resolve(rootDir, "src/routes"),
    "@tests": path.resolve(rootDir, "tests"),
    "@theme": path.resolve(rootDir, "src/theme"),
  };
};
