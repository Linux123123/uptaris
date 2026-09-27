import js from "@eslint/js";
import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["dist"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,tsx}"],
    languageOptions: { ecmaVersion: 2022, globals: globals.browser },
    plugins: { "react-hooks": reactHooks, "react-refresh": reactRefresh },
    rules: {
      ...reactHooks.configs.recommended.rules,
      // shadcn component files intentionally export variants and hooks beside components.
      "react-refresh/only-export-components": "off",
      "no-restricted-syntax": [
        "error",
        {
          selector: "ImportDeclaration[source.value='react'] > ImportNamespaceSpecifier",
          message: "Import React hooks and types by name.",
        },
        {
          selector: "MemberExpression[object.name='React'], TSQualifiedName[left.name='React']",
          message: "Use named React imports instead of React namespace members.",
        },
      ],
    },
  },
  {
    files: ["src/**/*.{ts,tsx}"],
    ignores: ["src/components/ui/**", "src/routeTree.gen.ts"],
    rules: {
      "padding-line-between-statements": [
        "error",
        { blankLine: "always", prev: "*", next: "return" },
        { blankLine: "always", prev: "*", next: ["function", "export"] },
        { blankLine: "always", prev: "*", next: ["if", "for", "while", "switch", "try"] },
        { blankLine: "always", prev: ["if", "for", "while", "switch", "try"], next: "*" },
        { blankLine: "always", prev: "import", next: "*" },
        { blankLine: "any", prev: "import", next: "import" },
      ],
    },
  },
);
