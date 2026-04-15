import pluginJs from "@eslint/js";
import eslintPluginImport from "eslint-plugin-import";
import eslintPluginPrettier from "eslint-plugin-prettier";
import pluginReact from "eslint-plugin-react";
import globals from "globals";
import tseslint from "typescript-eslint";

const importOrder = [
  "error",
  {
    "groups": [
      ["builtin", "external"],
      ["internal"],
      ["parent", "sibling", "index"]
    ],
    "pathGroups": [
      {
        pattern: "@{app,entities,features,shared,pages,widgets}/**",
        group: "internal",
        position: "before"
      }
    ],
    "pathGroupsExcludedImportTypes": ["builtin"],
    "newlines-between": "always",
    "alphabetize": { order: "asc", caseInsensitive: true }
  }
];

export default [
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
  { files: ["**/*.{js,mjs,cjs,ts,jsx,tsx}"] },
  {
    plugins: {
      import: eslintPluginImport,
      react: pluginReact,
      prettier: eslintPluginPrettier
    }
  },
  {
    languageOptions: { globals: globals.browser },
    rules: {
      "react/react-in-jsx-scope": 0,
      "comma-dangle": ["error", "never"],
      "import/order": importOrder
    }
  },
  {
    settings: {
      "import/resolver": {
        typescript: {},
        alias: {
          map: [
            ["@app", "./src/app"],
            ["@entities", "./src/entities"],
            ["@features", "./src/features"],
            ["@shared", "./src/shared"],
            ["@pages", "./src/pages"],
            ["@widgets", "./src/widgets"]
          ],
          extensions: [".js", ".jsx", ".ts", ".tsx"]
        }
      }
    }
  }
];
