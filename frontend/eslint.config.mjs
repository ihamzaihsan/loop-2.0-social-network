import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTypescript,
  {
    rules: {
      // API and WebSocket payloads are progressively typed at their boundaries.
      "@typescript-eslint/no-explicit-any": "off",
      // User-uploaded images are served dynamically by the authenticated Go API.
      "@next/next/no-img-element": "off",
      // These React Compiler rules are not required because the compiler is disabled.
      "react-hooks/immutability": "off",
      "react-hooks/set-state-in-effect": "off",
      // Page loaders intentionally key effects by route identifiers and session state.
      "react-hooks/exhaustive-deps": "off",
      "react/no-unescaped-entities": "off",
    },
  },
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
